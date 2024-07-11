// Provides the 'appcatalogs' command to export Giant Swarm app catalogs as Backstage entities.
package appcatalogs

import (
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/google/go-github/v63/github"
	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/githubrepo"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/helmrepoindex"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/teams"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/component"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/group"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/export"
)

const (
	// Name of the GitHub organization owning our teams and users.
	githubOrganization = "giantswarm"
)

var Command = &cobra.Command{
	Use:   "appcatalogs",
	Short: "Export Giant Swarm app catalogs as Backstage entities",
	Long: `The command takes a number of Giant Swarm app catalog URLs and exports one entity per unique app found.

Apps are deduplicated by name. Apps with the same name, even in different catalogs, are considered the same app.`,
	Run: runAppCatalogs,
}

const (
	// Catalogs
	clusterCatalogURL              = "https://giantswarm.github.io/cluster-catalog/index.yaml"
	controlPlaneCatalogURL         = "https://giantswarm.github.io/control-plane-catalog/index.yaml"
	defaultCatalogURL              = "https://giantswarm.github.io/default-catalog/index.yaml"
	giantSwarmAzureCatalogURL      = "https://giantswarm.github.io/giantswarm-azure-catalog/index.yaml"
	giantSwarmCatalogURL           = "https://giantswarm.github.io/giantswarm-catalog/index.yaml"
	giantSwarmPlaygroundCatalogURL = "https://giantswarm.github.io/giantswarm-playground-catalog/index.yaml"

	// Annotation keys
	teamAnnotation = "application.giantswarm.io/team"
)

var defaultCatalogURLs = []string{
	clusterCatalogURL,
	controlPlaneCatalogURL,
	defaultCatalogURL,
	giantSwarmAzureCatalogURL,
	giantSwarmCatalogURL,
	giantSwarmPlaygroundCatalogURL,
}

func init() {
	Command.PersistentFlags().StringSliceP("url", "", defaultCatalogURLs, "App catalog urls")
}

func runAppCatalogs(cmd *cobra.Command, args []string) {
	urls, err := cmd.PersistentFlags().GetStringSlice("url")
	if err != nil {
		log.Fatal(err)
	}

	path, err := cmd.Root().PersistentFlags().GetString("output")
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("Please set environment variable GITHUB_TOKEN to a personal GitHub access token (PAT).")
	}

	componentExporter := export.New(export.Config{TargetPath: path + "/components.yaml"})
	groupExporter := export.New(export.Config{TargetPath: path + "/groups.yaml"})

	teamsService, err := teams.New(teams.Config{
		GithubOrganization: githubOrganization,
		GithubAuthToken:    token,
	})
	if err != nil {
		log.Fatal(err)
	}

	entriesCount := 0
	apps := make(map[string]int)
	teamSlugCount := make(map[string]int)

	repoService, err := githubrepo.New(githubrepo.Config{
		GithubOrganization: githubOrganization,
		GithubAuthToken:    token,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Iterate over catalogs
	for _, url := range urls {
		fmt.Printf("Reading catalog %s\n", url)

		index, err := helmrepoindex.LoadFromURL(url)
		if err != nil {
			log.Fatalf("Error loading app catalog from %s: %s", url, err)
		}

		log.Printf("Catalog generated at %s", index.Generated)

		// Iterate over apps in catalog
		for appName := range index.Entries {
			entriesCount++
			if _, ok := apps[appName]; ok {
				// App already seen, skip
				continue
			}

			if len(index.Entries[appName]) < 1 {
				log.Printf("App %s has no releases", appName)
				continue
			}

			apps[appName]++

			component, err := componentFromCatalogEntry(index.Entries[appName][0], repoService)
			if err != nil {
				log.Printf("ERROR: Could not create component entity. %v", err)
				continue
			}

			teamSlugCount[component.Owner]++

			e := component.ToEntity()
			err = componentExporter.AddEntity(e)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

		}
	}

	log.Printf("Collected %d unique apps in %d entries", len(apps), entriesCount)

	// Collect group/team data
	slugs := make([]string, len(teamSlugCount))
	i := 0
	for s := range teamSlugCount {
		s = strings.TrimPrefix(s, "group:"+githubOrganization+"/")
		if s == "unspecified" || s == "" {
			continue
		}
		slugs[i] = s
		i++
	}

	err = componentExporter.WriteFile()
	if err != nil {
		log.Fatalf("ERROR writing components: %v", err)
	}
	log.Printf("Wrote file %s", componentExporter.TargetPath)

	log.Printf("Collected %d unique team slugs", len(slugs))
	slices.Sort(slugs)
	log.Printf("Team slugs: %s", strings.Join(slugs, " "))

	for _, slug := range slugs {
		if slug == "" {
			continue
		}

		team, err := teamsService.GetBySlug(slug)
		if err != nil {
			if e, ok := err.(*github.ErrorResponse); ok && e.Message == "Not Found" {
				log.Printf("ERROR: Team %q not found on GitHub", slug)
				continue
			} else {
				log.Fatalf("ERROR: %v", err)
			}
		}

		groupMembers, err := teamsService.GetMembers(team.GetSlug())
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		memberNames := make([]string, len(groupMembers))
		for i, member := range groupMembers {
			memberNames[i] = member.GetLogin()
		}

		group, err := groupFromTeam(team, memberNames)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}

		entity := group.ToEntity()
		err = groupExporter.AddEntity(entity)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
	}

	err = groupExporter.WriteFile()
	if err != nil {
		log.Fatalf("ERROR writing groups: %v", err)
	}
	log.Printf("Wrote file %s", groupExporter.TargetPath)
}

func groupFromTeam(team *github.Team, members []string) (*group.Group, error) {
	return group.New(team.GetSlug(),
		group.WithNamespace("giantswarm"),
		group.WithDescription(team.GetDescription()),
		group.WithTitle(team.GetName()),
		group.WithPictureURL(fmt.Sprintf("https://avatars.githubusercontent.com/t/%d?s=116&v=4", team.GetID())),
		group.WithMemberNames(members...),
	)
}

// Populates a catalog.Component from an helmrepoindex.Entry
func componentFromCatalogEntry(entry helmrepoindex.Entry, service *githubrepo.Service) (*component.Component, error) {
	// owner team
	team := "group:giantswarm/unspecified"
	if teamName, ok := entry.Annotations[teamAnnotation]; ok {
		// Adding Giant Swarm product team prefix
		if !strings.HasPrefix(teamName, "team-") {
			teamName = "team-" + teamName
		}

		team = "group:giantswarm/" + teamName
	}

	// release time
	releaseTime, err := time.Parse(time.RFC3339, entry.Created)
	if err != nil {
		return nil, err
	}

	githubSlug := detectGitHubSlug(&entry)
	if githubSlug == "" {
		return nil, fmt.Errorf("could not detect GitHub slug for app %s in app metadata", entry.Name)
	}

	repoDetails, err := service.GetDetails(entry.Name)
	if err != nil {
		return nil, err
	}

	// deployment names
	nameWithoutApp := strings.TrimSuffix(entry.Name, "-app")
	nameWithApp := nameWithoutApp + "-app"

	component, err := component.New(entry.Name,
		component.WithNamespace("giantswarm"),
		component.WithTitle(entry.Name),
		component.WithDescription(entry.Description),
		component.WithGithubProjectSlug(githubSlug),
		component.WithLatestReleaseTag(entry.Version),
		component.WithLatestReleaseTime(releaseTime),
		component.WithOwner(team),
		component.WithTags(entry.Keywords...),
		component.WithDeploymentNames(nameWithoutApp, nameWithApp),
		component.WithType("service"),
		component.WithHasReadme(true), // we assume all apps have a README
		component.WithDefaultBranch(repoDetails.DefaultBranch),
	)
	if err != nil {
		return nil, err
	}

	return component, nil
}

// Guessing the GitHub slug from the app metadata
//
// Note: This is highly Giant Swarm specific.
// Strategy:
// - Find a URL starting with "https://github.com/giantswarm/" in the URLs
// TODO:
// - try the app name appended to https://github.com/giantswarm/
// - try variations with/without -app suffix
func detectGitHubSlug(entry *helmrepoindex.Entry) string {
	prefix := "https://github.com/giantswarm/"

	if strings.HasPrefix(entry.Home, prefix) {
		return githubSlugFromURL(entry.Home)
	}

	for _, url := range entry.Sources {
		if strings.HasPrefix(url, prefix) {
			return githubSlugFromURL(url)
		}
	}

	for _, url := range entry.Urls {
		if strings.HasPrefix(url, prefix) {
			return githubSlugFromURL(url)
		}
	}

	return ""
}

func githubSlugFromURL(url string) string {
	slug := strings.TrimPrefix(url, "https://github.com/")
	slug = strings.TrimSuffix(slug, "/")
	return slug
}
