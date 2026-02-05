// Package crd provides the 'crd' command to export CRDs as Backstage API entities.
package crd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/crdconfig"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/githuburl"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/api"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/export"
)

// Command is the crd cobra command.
var Command = &cobra.Command{
	Use:   "crd <config-file>",
	Short: "Export CRDs as Backstage API entities",
	Long: `The command reads a YAML config file with CRD definitions and generates Backstage API entities.

The config file should be a YAML array with items containing:
  - url: GitHub URL to the CRD YAML file (required)
  - owner: Backstage owner reference (required)
  - lifecycle: Lifecycle stage (optional, defaults to "production")
  - system: System reference (optional)

Example config:
  - url: https://github.com/giantswarm/apiextensions-application/blob/main/config/crd/application.giantswarm.io_apps.yaml
    owner: group:default/team-honeybadger
    lifecycle: production
    system: app-platform

Use "-" as the config file path to read from stdin.

Arguments:
  config-file    Path to YAML config file, or "-" for stdin`,
	Args: cobra.ExactArgs(1),
	Run:  runCRD,
}

func init() {
	Command.PersistentFlags().StringP("namespace", "n", "default", "Backstage namespace for the API entities")
}

func runCRD(cmd *cobra.Command, args []string) {
	configPath := args[0]

	namespace, err := cmd.PersistentFlags().GetString("namespace")
	if err != nil {
		log.Fatal(err)
	}

	outputPath, err := cmd.Root().PersistentFlags().GetString("output")
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Println("WARN: GITHUB_TOKEN not set. Using unauthenticated requests (lower rate limits).")
	}

	// Create config service
	var configService *crdconfig.Service
	if configPath == "-" {
		// Read from stdin
		configService, err = crdconfig.New(crdconfig.Config{
			Reader: os.Stdin,
		})
	} else {
		configService, err = crdconfig.New(crdconfig.Config{
			FilePath: configPath,
		})
	}
	if err != nil {
		log.Fatalf("Failed to create config service: %v", err)
	}

	// Load config
	items, err := configService.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Found %d CRD definitions in config", len(items))

	// Create GitHub service
	githubService, err := githuburl.New(githuburl.Config{
		AuthToken: token,
	})
	if err != nil {
		log.Fatalf("Failed to create GitHub service: %v", err)
	}

	// Create exporter
	apiExporter := export.New(export.Config{TargetPath: outputPath + "/crds.yaml"})

	// Process each CRD
	numAPIs := 0
	for i, item := range items {
		log.Printf("[%d/%d] Processing CRD from: %s", i+1, len(items), item.URL)

		// Fetch CRD content
		content, err := githubService.FetchContent(item.URL)
		if err != nil {
			log.Printf("WARN: Failed to fetch CRD from %s: %v", item.URL, err)
			continue
		}

		// Parse CRD metadata
		crdMeta, err := githuburl.ParseCRDMetadata(content)
		if err != nil {
			log.Printf("WARN: Failed to parse CRD from %s: %v", item.URL, err)
			continue
		}

		// Build description
		description := crdMeta.Description
		if description == "" {
			description = fmt.Sprintf("Kubernetes Custom Resource Definition for %s", crdMeta.Kind)
		}

		// Create API entity
		apiEntity, err := api.New(
			crdMeta.Name,
			api.WithNamespace(namespace),
			api.WithTitle(crdMeta.Kind),
			api.WithDescription(description),
			api.WithOwner(item.Owner),
			api.WithLifecycle(item.Lifecycle),
			api.WithType("crd"),
			api.WithDefinition(content),
			api.WithSystem(item.System),
			api.WithTags("crd", "kubernetes"),
		)
		if err != nil {
			log.Printf("WARN: Failed to create API entity for %s: %v", crdMeta.Name, err)
			continue
		}

		// Add source annotation
		apiEntity.SetAnnotation("backstage.io/source-location", fmt.Sprintf("url:%s", item.URL))

		// Add CRD-specific annotations
		if crdMeta.Group != "" {
			apiEntity.SetAnnotation("giantswarm.io/crd-group", crdMeta.Group)
		}

		entity := apiEntity.ToEntity()
		if err := apiExporter.AddEntity(entity); err != nil {
			log.Fatalf("Error adding API entity: %v", err)
		}

		numAPIs++
		log.Printf("Created API entity: %s", crdMeta.Name)
	}

	// Write file
	if err := apiExporter.WriteFile(); err != nil {
		log.Fatalf("Error writing APIs file: %v", err)
	}

	fmt.Printf("\n%d API entities written to file %s with size %d bytes\n",
		numAPIs, apiExporter.TargetPath, apiExporter.Len())
}
