// Provides the 'groups' command to export Giant Swarm GitHub teams as Backstage group entities.
package groups

import (
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v88/github"
	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/teams"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/group"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/export"
)

// Name of the GitHub organization owning our teams and users.
const githubOrganization = "giantswarm"

var Command = &cobra.Command{
	Use:   "groups",
	Short: "Export groups catalog",
	Long:  `Exports Giant Swarm GitHub teams as Backstage group entities.`,
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("Please set environment variable GITHUB_TOKEN to a personal GitHub access token (PAT).")
	}

	path, err := cmd.Root().PersistentFlags().GetString("output")
	if err != nil {
		log.Fatalf("Error: could not access 'output' flag - %s", err)
	}

	teamsService, err := teams.New(teams.Config{
		GithubOrganization: githubOrganization,
		GithubAuthToken:    token,
	})
	if err != nil {
		log.Fatalf("Error: could not create teams service -- %v", err)
	}

	groupExporter := export.New(export.Config{TargetPath: path + "/groups.yaml"})

	teamsList, err := teamsService.GetAll()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Processing %d teams", len(teamsList))

	numGroups := 0
	for _, team := range teamsList {
		members, err := teamsService.GetMembers(team.GetSlug())
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		var memberNames []string
		for _, u := range members {
			memberNames = append(memberNames, u.GetLogin())
		}

		g, err := groupFromTeam(team, memberNames)
		if err != nil {
			log.Fatalf("Error: could not create group -- %v", err)
		}

		err = groupExporter.AddEntity(g.ToEntity())
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		numGroups++
	}

	err = groupExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error writing groups: %v", err)
	}

	fmt.Printf("\n%d groups written to file %s with size %d bytes\n", numGroups, groupExporter.TargetPath, groupExporter.Len())

	return nil
}

// groupFromTeam builds a Backstage group from a GitHub team and its member logins.
func groupFromTeam(team *github.Team, memberNames []string) (*group.Group, error) {
	parentTeamName := ""
	if team.GetParent() != nil {
		parentTeamName = team.GetParent().GetSlug()
	}

	return group.New(team.GetSlug(),
		group.WithTitle(team.GetName()),
		group.WithDescription(team.GetDescription()),
		group.WithPictureURL(fmt.Sprintf("https://avatars.githubusercontent.com/t/%d?s=116&v=4", team.GetID())),
		group.WithMemberNames(memberNames...),
		group.WithParentName(parentTeamName),
		group.WithGrafanaDashboardSelector(fmt.Sprintf("tags @> 'owner:%s'", team.GetSlug())),
	)
}
