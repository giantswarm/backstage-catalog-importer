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

const (
	teamsFlag  = "teams"
	parentFlag = "parent"
)

var Command = &cobra.Command{
	Use:   "groups",
	Short: "Export groups catalog",
	Long: `Exports Giant Swarm GitHub teams as Backstage group entities.

To avoid exposing sensitive team names (for example customer-specific teams), at
least one filter must be given:

  --teams   Comma-separated allowlist of team slugs to export. Only teams in this
            list are exported.
  --parent  Only export teams that are descendants of this team slug (for example
            "employees").

When both are given, a team is exported only if it is in the allowlist AND a
descendant of the parent team.`,
	RunE: run,
}

func init() {
	Command.Flags().StringSlice(teamsFlag, nil, "Allowlist of team slugs to export (comma-separated). Only these teams are exported.")
	Command.Flags().String(parentFlag, "", `Only export teams that are descendants of this parent team slug (e.g. "employees").`)
}

func run(cmd *cobra.Command, args []string) error {
	allowedTeams, err := cmd.Flags().GetStringSlice(teamsFlag)
	if err != nil {
		return err
	}
	parent, err := cmd.Flags().GetString(parentFlag)
	if err != nil {
		return err
	}

	// Require an explicit filter to avoid accidentally exporting all teams,
	// which may expose sensitive (e.g. customer) team names.
	if len(allowedTeams) == 0 && parent == "" {
		log.Fatalf("Error: refusing to export all teams. Specify --%s and/or --%s to select which teams to expose.", teamsFlag, parentFlag)
	}

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

	teamsList, err := teamsService.GetAll()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Found %d teams in total", len(teamsList))

	// Map of team slug -> parent slug, used to resolve team ancestry.
	parentBySlug := make(map[string]string, len(teamsList))
	for _, t := range teamsList {
		parentBySlug[t.GetSlug()] = t.GetParent().GetSlug()
	}

	allowSet := make(map[string]bool, len(allowedTeams))
	for _, s := range allowedTeams {
		allowSet[s] = true
	}

	groupExporter := export.New(export.Config{TargetPath: path + "/groups.yaml"})

	numGroups := 0
	exported := make(map[string]bool)
	for _, team := range teamsList {
		slug := team.GetSlug()

		if len(allowSet) > 0 && !allowSet[slug] {
			continue
		}
		if parent != "" && !isDescendant(slug, parent, parentBySlug) {
			continue
		}

		members, err := teamsService.GetMembers(slug)
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
		exported[slug] = true
		numGroups++
	}

	// Warn about allowlisted teams that were not exported, which usually means a
	// typo in the slug or exclusion by the --parent filter.
	for _, s := range allowedTeams {
		if !exported[s] {
			log.Printf("WARN: allowlisted team %q was not exported (not found or excluded by --%s)", s, parentFlag)
		}
	}

	err = groupExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error writing groups: %v", err)
	}

	fmt.Printf("\n%d groups written to file %s with size %d bytes\n", numGroups, groupExporter.TargetPath, groupExporter.Len())

	return nil
}

// isDescendant reports whether the team identified by slug is a (transitive)
// descendant of ancestor, following the parent chain in parentBySlug. The
// ancestor team itself is not considered its own descendant. It guards against
// cycles in the (otherwise tree-shaped) hierarchy.
func isDescendant(slug, ancestor string, parentBySlug map[string]string) bool {
	seen := map[string]bool{slug: true}
	for cur := parentBySlug[slug]; cur != ""; cur = parentBySlug[cur] {
		if cur == ancestor {
			return true
		}
		if seen[cur] {
			break
		}
		seen[cur] = true
	}
	return false
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
