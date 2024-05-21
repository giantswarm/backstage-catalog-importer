package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/pkg/appcatalog"
	"github.com/giantswarm/backstage-catalog-importer/pkg/catalog"
	"github.com/giantswarm/backstage-catalog-importer/pkg/export"
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
func componentFromCatalogEntry(entry appcatalog.Entry) (*catalog.Component, error) {
	// owner team
	team := "unspecified"
	if val, ok := entry.Annotations[teamAnnotation]; ok {
		team = val
		// Adding Giant Swarm product team prefix
		if !strings.HasPrefix(team, "team-") {
			team = "team-" + team
		}
	}

	// release time
	releaseTime, err := time.Parse(time.RFC3339, entry.Created)
	if err != nil {
		return nil, err
	}

	// source URL
	// Note: This is very Giant Swarm specific. We assume `https://github.com/` as
	// the host and "giantswarm" as the organization name. This works for all
	// Giant Swarm catalogs, but will not work for customer catalogs.
	githubSlug := "giantswarm/" + entry.Name

	component, err := catalog.NewComponent(entry.Name,
		catalog.WithNamespace("giantswarm"),
		catalog.WithTitle(entry.Name),
		catalog.WithDescription(entry.Description),
		catalog.WithGithubProjectSlug(githubSlug),
		catalog.WithLatestReleaseTag(entry.Version),
		catalog.WithLatestReleaseTime(releaseTime),
		catalog.WithOwner(team),
		catalog.WithTags(entry.Keywords...),
		catalog.WithType("service"),
		catalog.WithHasReadme(true), // we assume all apps have a README
	)
	if err != nil {
		return nil, err
	}

	return component, nil
}
