// Provides the 'users' command to export a users catalog.
package users

import (
	"context"
	"log"
	"os"
	"sort"

	"github.com/google/go-github/v65/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/user"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/export"
)

var Command = &cobra.Command{
	Use:   "users",
	Short: "Export users catalog",
	Long:  `Exports Giant Swarm users, either for customer catalogs or for GS-internal use.`,
	RunE:  run,
}

const (
	internalFlag = "internal"
	orgFlag      = "org"
	outputFlag   = "output"
)

func init() {
	Command.Flags().BoolP(internalFlag, "i", false, "Create a Giant Swarm internal catalog, which includes email adresses.")
	Command.Flags().String(orgFlag, "giantswarm", "GitHub organization to export users from")

	Command.PersistentFlags().StringP(outputFlag, "o", ".", "Output directory path")
}

func run(cmd *cobra.Command, args []string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("Please set environment variable GITHUB_TOKEN to a personal GitHub access token (PAT).")
	}

	internal, err := cmd.Flags().GetBool(internalFlag)
	if err != nil {
		return err
	}

	path, err := cmd.PersistentFlags().GetString(outputFlag)
	if err != nil {
		log.Fatalf("Error: could not access 'output' flag - %s", err)
	}

	org, err := cmd.Flags().GetString(orgFlag)
	if err != nil {
		log.Fatalf("Error: could not access 'org' flag - %s", err)
	}

	// Github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	userExporter := export.New(export.Config{TargetPath: path + "/users.yaml"})
	numUsers := 0

	opt := &github.ListMembersOptions{
		PublicOnly:  false,
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allMembers []*github.User
	for {
		members, resp, err := client.Organizations.ListMembers(ctx, org, opt)
		if err != nil {
			log.Fatalf("Error: could not list organization members -- %v", err)
		}
		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	// Sort by user name
	sort.Slice(allMembers, func(i, j int) bool {
		return allMembers[i].GetLogin() < allMembers[j].GetLogin()
	})

	for _, githubUser := range allMembers {
		detailedUser, _, err := client.Users.Get(ctx, githubUser.GetLogin())
		if err != nil {
			log.Fatalf("Error: could not read detailed user entry -- %v", err)
		}

		user, err := user.New(githubUser.GetLogin(),
			user.WithDisplayName(detailedUser.GetName()),
			user.WithPictureURL(detailedUser.GetAvatarURL()),
			user.WithDescription(detailedUser.GetBio()),
		)
		if err != nil {
			log.Fatalf("Error: could not create user - %v", err)
		}

		if internal {
			// Email will only be published internally at Giant Swarm
			user.Email = detailedUser.GetEmail()
		} else {
			// In customer catalogs, Giant Swarm entities use the "giantswarm" namespace
			user.Namespace = "giantswarm"
		}

		err = userExporter.AddEntity(user.ToEntity())
		if err != nil {
			log.Fatalf("Error: could not add user entity - %v", err)
		}
		numUsers++
	}

	err = userExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error: could not write users -- %v", err)
	}

	return nil
}
