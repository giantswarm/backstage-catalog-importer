// Provides the 'charts' command to export OCI registry charts as Backstage entities.
package charts

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/ociregistry"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/component"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/export"
)

// exportStats tracks statistics about exported component entities.
type exportStats struct {
	total            int
	annotationCounts map[string]int
}

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

const (
	audienceOciAnnotation          = "io.giantswarm.application.audience"
	audienceLegacyChartAnnnotation = "application.giantswarm.io/audience"
	audienceBackstageAnnotation    = audienceLegacyChartAnnnotation

	managedOciAnnotation         = "io.giantswarm.application.managed"
	managedLegacyChartAnnotation = "application.giantswarm.io/managed"
	managedBackstageAnnotation   = managedLegacyChartAnnotation

	teamOciAnnotation         = "io.giantswarm.application.team"
	teamLegacyChartAnnotation = "application.giantswarm.io/team"

	iconBackstageAnnotation = "giantswarm.io/icon-url"

	registryBackstageAnnotation   = "giantswarm.io/oci-registry"
	repositoryBackstageAnnotation = "giantswarm.io/oci-repository"
	tagBackstageAnnotation        = "giantswarm.io/oci-tag"

	audienceAll        = "all"
	audienceGiantSwarm = "giantswarm"
)

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
	stats := &exportStats{
		annotationCounts: make(map[string]int),
	}

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

		// Track statistics
		stats.total++
		stats.trackAnnotations(entity.Metadata.Annotations)

		log.Printf("Created component: %s", comp.Name)
	}

	// Write the components file
	err = componentExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error writing components file: %v", err)
	}

	fmt.Printf("\n%d components written to file %s with size %d bytes\n",
		stats.total, componentExporter.TargetPath, componentExporter.Len())

	// Print statistics report
	stats.printReport()
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

	// Extract version, creation time, icon, team owner, and GitHub project slug
	version := tag
	createdTime := time.Now() // Default to now if we can't extract creation time
	var iconURL string
	var githubProjectSlug string
	componentOwner := fmt.Sprintf("group:%s/unspecified", namespace) // Default owner based on namespace

	// See https://github.com/giantswarm/roadmap/issues/4156#issuecomment-3589340419
	managed := false
	audience := audienceAll

	if configMap != nil {
		// Extract version from top-level (Helm chart config structure)
		if ver, ok := configMap["version"].(string); ok && ver != "" {
			version = ver
		}

		// Extract icon from top-level (Helm chart config structure)
		if icon, ok := configMap["icon"].(string); ok && icon != "" {
			iconURL = icon
		}

		// Extract GitHub project slug from home field (Helm chart config structure)
		// Expected format: https://github.com/giantswarm/repository-name
		if home, ok := configMap["home"].(string); ok && home != "" {
			const githubGiantSwarmPrefix = "https://github.com/giantswarm/"
			if strings.HasPrefix(home, githubGiantSwarmPrefix) {
				// Extract "giantswarm/repository-name" from the URL
				githubProjectSlug = "giantswarm/" + strings.TrimPrefix(home, githubGiantSwarmPrefix)
			}
		}
		if githubProjectSlug == "" {
			log.Printf("WARN: 'home' property is not a valid GitHub URL for %s:%s", repo, tag)
		}

		// Extract team owner from annotations (Helm chart config structure)
		// The annotations field contains Giant Swarm specific metadata
		if annotations, ok := configMap["annotations"].(map[string]interface{}); ok {
			if team, ok := annotations[teamOciAnnotation].(string); ok && team != "" {
				componentOwner = formatTeamOwner(team, namespace)
			} else if team, ok := annotations[teamLegacyChartAnnotation].(string); ok && team != "" {
				componentOwner = formatTeamOwner(team, namespace)
			}

			if val, exists := annotations[managedOciAnnotation]; exists {
				if managedValue, ok := val.(bool); ok {
					managed = managedValue
				} else {
					log.Printf("WARN: '%s' annotation value is not a boolean for %s:%s", managedOciAnnotation, repo, tag)
				}
			} else if val, exists := annotations[managedLegacyChartAnnotation]; exists {
				if managedValue, ok := val.(bool); ok {
					managed = managedValue
				} else {
					log.Printf("WARN: '%s' annotation value is not a boolean for %s:%s", managedLegacyChartAnnotation, repo, tag)
				}
			}

			if audienceValue, ok := annotations[audienceOciAnnotation].(string); ok {
				audience = audienceValue
			} else if audienceValue, ok := annotations[audienceLegacyChartAnnnotation].(string); ok {
				audience = audienceValue
			}
			if audience != audienceAll && audience != audienceGiantSwarm {
				log.Printf("WARN: audience annotation value '%s' is not a valid audience for %s:%s", audience, repo, tag)
				audience = audienceAll // back to default
			}
		}

		// Try to extract creation time
		if created, ok := configMap["created"].(string); ok {
			if t, err := time.Parse(time.RFC3339, created); err == nil {
				createdTime = t
			}
		}
	}

	// Build component options
	componentOpts := []component.Option{
		component.WithNamespace(namespace),
		component.WithTitle(name),
		component.WithDescription(description),
		component.WithOwner(componentOwner),
		component.WithType(componentType),
		component.WithLatestReleaseTag(version),
		component.WithLatestReleaseTime(createdTime),
		component.WithTags("oci", "helm-chart"),
	}

	// Add GitHub project slug if available
	if githubProjectSlug != "" {
		componentOpts = append(componentOpts, component.WithGithubProjectSlug(githubProjectSlug))
	}

	// Create the component
	comp, err := component.New(name, componentOpts...)
	if err != nil {
		return nil, err
	}

	// Add OCI-specific annotations
	comp.SetAnnotation(registryBackstageAnnotation, registryHostname)
	comp.SetAnnotation(repositoryBackstageAnnotation, repo)
	comp.SetAnnotation(tagBackstageAnnotation, tag)
	comp.SetAnnotation(audienceBackstageAnnotation, audience)
	comp.SetAnnotation(managedBackstageAnnotation, strconv.FormatBool(managed))

	// Add icon URL if available
	if iconURL != "" {
		comp.SetAnnotation(iconBackstageAnnotation, iconURL)
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

// trackAnnotations counts each annotation key present in the entity.
func (s *exportStats) trackAnnotations(annotations map[string]string) {
	for key := range annotations {
		s.annotationCounts[key]++
	}
}

// printReport prints a summary of annotation coverage statistics.
func (s *exportStats) printReport() {
	if s.total == 0 {
		return
	}

	fmt.Println()
	fmt.Println("=== Annotation Statistics ===")
	fmt.Println()

	// Sort annotations by count (descending), then alphabetically
	type annotationCount struct {
		key   string
		count int
	}
	var annotations []annotationCount
	for key, count := range s.annotationCounts {
		annotations = append(annotations, annotationCount{key, count})
	}
	sort.Slice(annotations, func(i, j int) bool {
		if annotations[i].count != annotations[j].count {
			return annotations[i].count > annotations[j].count
		}
		return annotations[i].key < annotations[j].key
	})

	for _, ac := range annotations {
		pct := float64(ac.count) * 100 / float64(s.total)
		fmt.Printf("  %-45s %3d / %d  (%5.1f%%)\n", ac.key+":", ac.count, s.total, pct)
	}
	fmt.Println()
}
