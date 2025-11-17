// Provides the 'users' command to export a users catalog.
package users

import (
	"context"
	"log"
	"os"

	"github.com/google/go-github/v79/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/personio"
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
	outputFlag   = "output"
)

func init() {
	Command.Flags().BoolP(internalFlag, "i", false, "Create a Giant Swarm internal catalog, which includes email adresses.")
	Command.PersistentFlags().StringP(outputFlag, "o", ".", "Output directory path")
}

func run(cmd *cobra.Command, args []string) error {
	// Personio credentials
	personioClientID := os.Getenv("PERSONIO_CLIENT_ID")
	if personioClientID == "" {
		log.Fatal("Please set environment variable PERSONIO_CLIENT_ID to the Personio client ID.")
	}
	personioClientSecret := os.Getenv("PERSONIO_CLIENT_SECRET")
	if personioClientSecret == "" {
		log.Fatal("Please set environment variable PERSONIO_CLIENT_SECRET to the Personio client secret.")
	}

	// GitHub credentials
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

	ctx := context.Background()

	// Github client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)

	userExporter := export.New(export.Config{TargetPath: path + "/users.yaml"})

	employees, err := personio.GetActiveEmployees(ctx, personioClientID, personioClientSecret)
	if err != nil {
		log.Fatalf("Error: could not get employees from Personio -- %v", err)
	}

	for _, employee := range employees {
		// Get Github details for each employee
		if employee.GithubHandle == "" {
			// If the user has no GitHub handle, we skip them
			log.Printf("Warning: no GitHub handle found for %s %s (%s) -- skipping", employee.FirstName, employee.LastName, employee.Email)
			continue
		}
		githubDetails, _, err := githubClient.Users.Get(ctx, employee.GithubHandle)
		if err != nil {
			log.Fatalf("Error: could not read detailed user entry -- %v", err)
		}

		user, err := user.New(employee.GithubHandle,
			user.WithPictureURL(githubDetails.GetAvatarURL()),
			user.WithDescription(githubDetails.GetBio()),
			user.WithEmail(employee.Email),
			user.WithGitHubHandle(employee.GithubHandle),
			user.WithGitHubID(githubDetails.GetID()),
		)
		if err != nil {
			log.Fatalf("Error: could not create user - %v", err)
		}

		if internal {
			user.DisplayName = employee.FirstName + " " + employee.LastName
		} else {
			// In customer catalogs, Giant Swarm entities use the "giantswarm" namespace
			user.Namespace = "giantswarm"
			user.DisplayName = githubDetails.GetName()
		}

		err = userExporter.AddEntity(user.ToEntity())
		if err != nil {
			log.Fatalf("Error: could not add user entity - %v", err)
		}
	}

	err = userExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error: could not write users -- %v", err)
	}

	return nil
}
