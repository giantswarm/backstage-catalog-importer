package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/pkg/catalog/component"
	"github.com/giantswarm/backstage-catalog-importer/pkg/export"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/appcatalog"
)

var appCatalogsCmd = &cobra.Command{
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
	appCatalogsCmd.PersistentFlags().StringSliceP("url", "", defaultCatalogURLs, "App catalog urls")
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

	componentExporter := export.New(export.Config{TargetPath: path + "/components.yaml"})

	entriesCount := 0
	apps := make(map[string]int)

	// Iterate over catalogs
	for _, url := range urls {
		fmt.Printf("Reading catalog %s\n", url)

		index, err := appcatalog.LoadFromURL(url)
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

			component, err := componentFromCatalogEntry(index.Entries[appName][0])
			if err != nil {
				log.Printf("ERROR: Could not create component entity. %v", err)
				continue
			}

			e := component.ToEntity()
			err = componentExporter.AddEntity(e)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

		}
	}

	log.Printf("Collected %d unique apps in %d entries", len(apps), entriesCount)

	err = componentExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error writing components: %v", err)
	}
	log.Printf("Wrote file %s", componentExporter.TargetPath)
}

// Populates a catalog.Component from an appcatalog.Entry
func componentFromCatalogEntry(entry appcatalog.Entry) (*component.Component, error) {
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

	component, err := component.New(entry.Name,
		component.WithNamespace("giantswarm"),
		component.WithTitle(entry.Name),
		component.WithDescription(entry.Description),
		component.WithGithubProjectSlug(githubSlug),
		component.WithLatestReleaseTag(entry.Version),
		component.WithLatestReleaseTime(releaseTime),
		component.WithOwner(team),
		component.WithTags(entry.Keywords...),
		component.WithType("service"),
		component.WithHasReadme(true), // we assume all apps have a README
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
func detectGitHubSlug(entry *appcatalog.Entry) string {
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
