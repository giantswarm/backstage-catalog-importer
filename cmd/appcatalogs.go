package cmd

import (
	"fmt"
	"log"

	"github.com/giantswarm/backstage-catalog-importer/pkg/appcatalog"
	"github.com/giantswarm/backstage-catalog-importer/pkg/catalog"
	"github.com/giantswarm/backstage-catalog-importer/pkg/export"
	"github.com/spf13/cobra"
)

var appCatalogsCmd = &cobra.Command{
	Use:   "appcatalogs",
	Short: "Export Giant Swarm app catalogs as Backstage entities",
	Long: `The command takes a number of Giant Swarm app catalog URLs and exports one entity per unique app found.

Apps are deduplicated by name. Apps with the same name, even in different catalogs, are considered the same app.`,
	Run: runAppCatalogs,
}

const (
	clusterCatalogURL              = "https://giantswarm.github.io/cluster-catalog/index.yaml"
	controlPlaneCatalogURL         = "https://giantswarm.github.io/control-plane-catalog/index.yaml"
	defaultCatalogURL              = "https://giantswarm.github.io/default-catalog/index.yaml"
	giantSwarmAzureCatalogURL      = "https://giantswarm.github.io/giantswarm-azure-catalog/index.yaml"
	giantSwarmCatalogURL           = "https://giantswarm.github.io/giantswarm-catalog/index.yaml"
	giantSwarmPlaygroundCatalogURL = "https://giantswarm.github.io/giantswarm-playground-catalog/index.yaml"
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

	// TODO: create exporter
	componentExporter := export.New(export.Config{TargetPath: path + "/components.yaml"})

	entriesCount := 0
	apps := make(map[string]int)

	for _, url := range urls {
		fmt.Printf("Reading catalog %s\n", url)

		index, err := appcatalog.LoadFromURL(url)
		if err != nil {
			log.Fatalf("Error loading app catalog from %s: %s", url, err)
		}

		log.Printf("Catalog generated at %s", index.Generated)

		for appName := range index.Entries {
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
				log.Fatalf("Error: %v", err)
			}

			e := component.ToEntity()
			err = componentExporter.AddEntity(e)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

		}
	}

	err = componentExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error writing components: %v", err)
	}

	log.Printf("Found %d unique apps in %d entries", len(apps), entriesCount)
	log.Printf("TODO: Export result to %s", path)
}

// Populates a catalog.Component from an appcatalog.Entry
func componentFromCatalogEntry(entry appcatalog.Entry) (*catalog.Component, error) {
	component, err := catalog.NewComponent(entry.Name,
		catalog.WithDescription(entry.Description),
		catalog.WithLatestReleaseTag(entry.Version),
		catalog.WithTags(entry.Keywords...),
		//catalog.WithLatestReleaseTime(entry.Created), TODO: parse time
		// TODO: pass github project slug from "home" field
		// TODO: pass github team slug from "application.giantswarm.io/team" annotation
	)
	if err != nil {
		return nil, err
	}

	return component, nil
}
