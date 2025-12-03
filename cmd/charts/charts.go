// Provides the 'charts' command to export OCI registry charts as Backstage entities.
package charts

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/ociregistry"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/component"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/export"
)

var Command = &cobra.Command{
	Use:   "charts <registry>",
	Short: "Export OCI registry charts as Backstage entities",
	Long: `The command connects to an OCI registry and exports Helm charts as Backstage component entities.

Charts are discovered by listing repositories with a specified prefix and extracting metadata from their manifests.

Arguments:
  registry    OCI registry hostname (e.g., gsoci.azurecr.io)`,
	Args: cobra.ExactArgs(1),
	Run:  runCharts,
}

func init() {
	Command.PersistentFlags().StringP("prefix", "p", "", "Repository prefix to filter charts (optional)")
	Command.PersistentFlags().StringP("namespace", "n", "default", "Backstage namespace for the components")
	Command.PersistentFlags().StringP("type", "t", "service", "Component type")
}

func runCharts(cmd *cobra.Command, args []string) {
	// Get registry hostname from positional argument
	registryHostname := args[0]

	prefix, err := cmd.PersistentFlags().GetString("prefix")
	if err != nil {
		log.Fatal(err)
	}

	namespace, err := cmd.PersistentFlags().GetString("namespace")
	if err != nil {
		log.Fatal(err)
	}

	componentType, err := cmd.PersistentFlags().GetString("type")
	if err != nil {
		log.Fatal(err)
	}

	outputPath, err := cmd.Root().PersistentFlags().GetString("output")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Create OCI registry client
	registry, err := ociregistry.NewRegistry(ctx, ociregistry.Config{
		Hostname: registryHostname,
	})
	if err != nil {
		log.Fatalf("Failed to create OCI registry client: %v", err)
	}

	log.Printf("Connected to OCI registry: %s", registryHostname)

	// List repositories
	repositories, err := registry.ListRepositories(ctx, prefix)
	if err != nil {
		log.Fatalf("Failed to list repositories: %v", err)
	}

	log.Printf("Found %d repositories with prefix '%s'", len(repositories), prefix)

	componentExporter := export.New(export.Config{TargetPath: outputPath + "/charts.yaml"})
	componentsCreated := 0

	// Process each repository
	for _, repo := range repositories {
		log.Printf("Processing repository: %s", repo)

		// List tags for the repository
		tags, err := registry.ListRepositoryTags(ctx, repo)
		if err != nil {
			log.Printf("WARN: Failed to list tags for repository %s: %v", repo, err)
			continue
		}

		if len(tags) == 0 {
			log.Printf("WARN: No tags found for repository %s", repo)
			continue
		}

		// Use the first tag (typically latest or most recent)
		tag := tags[0]
		log.Printf("Using tag: %s", tag)

		// Get manifest for metadata extraction
		configMap, err := registry.GetRepositoryManifest(ctx, repo, tag)
		if err != nil {
			log.Printf("WARN: Failed to get manifest for %s:%s: %v", repo, tag, err)
			continue
		}

		// Create component from repository and manifest data
		comp, err := createComponentFromOCIChart(repo, tag, configMap, namespace, componentType, registryHostname)
		if err != nil {
			log.Printf("WARN: Failed to create component for %s:%s: %v", repo, tag, err)
			continue
		}

		entity := comp.ToEntity()
		err = componentExporter.AddEntity(entity)
		if err != nil {
			log.Fatalf("Error adding component entity: %v", err)
		}

		componentsCreated++
		log.Printf("Created component: %s", comp.Name)
	}

	// Write the components file
	err = componentExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error writing components file: %v", err)
	}

	fmt.Printf("\n%d components written to file %s with size %d bytes\n",
		componentsCreated, componentExporter.TargetPath, componentExporter.Len())
}

// createComponentFromOCIChart creates a Backstage component from OCI chart metadata
func createComponentFromOCIChart(repo string, tag string, configMap map[string]interface{}, namespace, componentType, registryHostname string) (*component.Component, error) {
	// Extract component name from repository path
	// Remove any org prefix and clean up the name
	name := repo
	if strings.Contains(repo, "/") {
		parts := strings.Split(repo, "/")
		name = parts[len(parts)-1]
	}

	// Extract description from config if available
	// Helm chart configs have description at the top level
	description := fmt.Sprintf("OCI chart from %s", repo)
	if configMap != nil {
		if desc, ok := configMap["description"].(string); ok && desc != "" {
			description = desc
		}
	}

	// Extract version, creation time, icon, and team owner
	version := tag
	createdTime := time.Now() // Default to now if we can't extract creation time
	var iconURL string
	componentOwner := fmt.Sprintf("group:%s/unspecified", namespace) // Default owner based on namespace

	if configMap != nil {
		// Extract version from top-level (Helm chart config structure)
		if ver, ok := configMap["version"].(string); ok && ver != "" {
			version = ver
		}

		// Extract icon from top-level (Helm chart config structure)
		if icon, ok := configMap["icon"].(string); ok && icon != "" {
			iconURL = icon
		}

		// Extract team owner from annotations (Helm chart config structure)
		// The annotations field contains Giant Swarm specific metadata
		if annotations, ok := configMap["annotations"].(map[string]interface{}); ok {
			if team, ok := annotations["application.giantswarm.io/team"].(string); ok && team != "" {
				componentOwner = formatTeamOwner(team, namespace)
			}
		}

		// Try to extract creation time
		if created, ok := configMap["created"].(string); ok {
			if t, err := time.Parse(time.RFC3339, created); err == nil {
				createdTime = t
			}
		}
	}

	// Create the component
	comp, err := component.New(name,
		component.WithNamespace(namespace),
		component.WithTitle(name),
		component.WithDescription(description),
		component.WithOwner(componentOwner),
		component.WithType(componentType),
		component.WithLatestReleaseTag(version),
		component.WithLatestReleaseTime(createdTime),
		component.WithTags("oci", "helm-chart"),
	)
	if err != nil {
		return nil, err
	}

	// Add OCI-specific annotations
	comp.SetAnnotation("giantswarm.io/oci-registry", registryHostname)
	comp.SetAnnotation("giantswarm.io/oci-repository", repo)
	comp.SetAnnotation("giantswarm.io/oci-tag", tag)

	// Add icon URL if available
	if iconURL != "" {
		comp.SetAnnotation("giantswarm.io/icon-url", iconURL)
	}

	return comp, nil
}

// formatTeamOwner formats a team name to the proper Backstage owner format using the given namespace
// Examples:
//   - "honeybadger", "giantswarm" -> "group:giantswarm/team-honeybadger"
//   - "team-atlas", "default" -> "group:default/team-atlas"
//   - "team-bigmac", "production" -> "group:production/team-bigmac"
func formatTeamOwner(team, namespace string) string {
	// Ensure the team name has the "team-" prefix
	if !strings.HasPrefix(team, "team-") {
		team = "team-" + team
	}

	// Return the properly formatted owner string with the given namespace
	return fmt.Sprintf("group:%s/%s", namespace, team)
}
