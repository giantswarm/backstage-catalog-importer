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
	componentutil "github.com/giantswarm/backstage-catalog-importer/pkg/util/component"
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

Only charts with the annotation io.giantswarm.application.audience set to "all" in the config blob are included in the output.

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

	audienceAll        = "all"
	audienceGiantSwarm = "giantswarm"
)

func init() {
	Command.PersistentFlags().StringP("prefix", "p", "", "Repository prefix to filter charts (optional)")
	Command.PersistentFlags().StringP("namespace", "n", "default", "Backstage namespace for the components")
	Command.PersistentFlags().StringP("type", "t", "service", "Component type")
	Command.PersistentFlags().IntP("limit", "l", 0, "Limit the number of charts to process (0 = no limit, for testing)")
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

	limit, err := cmd.PersistentFlags().GetInt("limit")
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

	// Apply limit if specified
	if limit > 0 && len(repositories) > limit {
		log.Printf("Limiting to first %d repositories (--limit flag)", limit)
		repositories = repositories[:limit]
	}

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
		manifestInfo, err := registry.GetRepositoryManifest(ctx, repo, tag)
		if err != nil {
			log.Printf("WARN: Failed to get manifest for %s:%s: %v", repo, tag, err)
			continue
		}

		// Filter charts: only include charts with audience annotation set to "all"
		if !shouldIncludeChart(manifestInfo.Config) {
			log.Printf("Skipping chart %s:%s (audience annotation is not 'all')", repo, tag)
			continue
		}

		// Create component from repository and manifest data
		comp, err := createComponentFromOCIChart(repo, tag, manifestInfo, namespace, componentType, registryHostname)
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
func createComponentFromOCIChart(repo string, tag string, manifestInfo *ociregistry.ManifestInfo, namespace, componentType, registryHostname string) (*component.Component, error) {
	configMap := manifestInfo.Config

	// Extract GitHub project slug and repository name from home field (Helm chart config structure)
	// Expected format: https://github.com/giantswarm/repository-name
	var githubProjectSlug string
	var githubRepoName string

	if configMap != nil {
		if home, ok := configMap["home"].(string); ok && home != "" {
			const githubGiantSwarmPrefix = "https://github.com/giantswarm/"
			if strings.HasPrefix(home, githubGiantSwarmPrefix) {
				// Extract repository name from the URL and remove any trailing slashes
				repoNameFromURL := strings.TrimPrefix(home, githubGiantSwarmPrefix)
				repoNameFromURL = strings.TrimSuffix(repoNameFromURL, "/")

				// Build project slug and repository name
				githubProjectSlug = "giantswarm/" + repoNameFromURL
				githubRepoName = repoNameFromURL
			}
		}
	}

	// We require a valid GitHub repository to create the component
	if githubProjectSlug == "" {
		return nil, fmt.Errorf("cannot match chart to GitHub repository: 'home' field is missing or not a valid GitHub URL")
	}

	// Use the GitHub repository name as the component name
	name := githubRepoName

	// Extract description from config if available
	// Helm chart configs have description at the top level
	description := fmt.Sprintf("OCI chart from %s", repo)
	if configMap != nil {
		if desc, ok := configMap["description"].(string); ok && desc != "" {
			description = desc
		}
	}

	// Extract version, appVersion, creation time, icon, team owner, and chart type
	chartVersion := tag
	var appVersion string
	var chartType string
	createdTime := time.Time{} // Zero value, will be set from manifest annotations
	var iconURL string
	componentOwner := fmt.Sprintf("group:%s/unspecified", namespace) // Default owner based on namespace

	// See https://github.com/giantswarm/roadmap/issues/4156#issuecomment-3589340419
	managed := false
	audience := audienceAll

	// Extract creation time from OCI manifest annotations
	if manifestInfo.Annotations != nil {
		if created, ok := manifestInfo.Annotations["org.opencontainers.image.created"]; ok && created != "" {
			if t, err := time.Parse(time.RFC3339, created); err == nil {
				createdTime = t
			} else {
				log.Printf("WARN: Failed to parse org.opencontainers.image.created annotation '%s' for %s:%s: %v", created, repo, tag, err)
			}
		}
	}

	if configMap != nil {
		// Extract version from top-level (Helm chart config structure)
		if ver, ok := configMap["version"].(string); ok && ver != "" {
			chartVersion = ver
		}

		// Extract appVersion from top-level (Helm chart config structure)
		if appVer, ok := configMap["appVersion"].(string); ok && appVer != "" {
			appVersion = appVer
		}

		// Extract chart type from top-level (Helm chart config structure)
		if cType, ok := configMap["type"].(string); ok {
			chartType = cType
		}

		// Extract icon from top-level (Helm chart config structure)
		if icon, ok := configMap["icon"].(string); ok && icon != "" {
			iconURL = icon
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
				if strVal, ok := val.(string); ok {
					if managedValue, err := strconv.ParseBool(strVal); err == nil {
						managed = managedValue
					} else {
						log.Printf("WARN: '%s' annotation value '%s' is not a valid boolean (expected 'true' or 'false') for %s:%s", managedOciAnnotation, strVal, repo, tag)
					}
				} else {
					log.Printf("WARN: '%s' annotation value is not a string for %s:%s", managedOciAnnotation, repo, tag)
				}
			} else if val, exists := annotations[managedLegacyChartAnnotation]; exists {
				if strVal, ok := val.(string); ok {
					if managedValue, err := strconv.ParseBool(strVal); err == nil {
						managed = managedValue
					} else {
						log.Printf("WARN: '%s' annotation value '%s' is not a valid boolean (expected 'true' or 'false') for %s:%s", managedLegacyChartAnnotation, strVal, repo, tag)
					}
				} else {
					log.Printf("WARN: '%s' annotation value is not a string for %s:%s", managedLegacyChartAnnotation, repo, tag)
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
	}

	// Prepare deployment names using shared utility function
	deploymentNames := componentutil.GenerateDeploymentNames(name)

	// Build component options
	componentOpts := []component.Option{
		component.WithNamespace(namespace),
		component.WithTitle(name),
		component.WithDescription(description),
		component.WithOwner(componentOwner),
		component.WithType(componentType),
		component.WithLatestReleaseTag(chartVersion),
		component.WithLatestReleaseTime(createdTime),
		component.WithDeploymentNames(deploymentNames...),
		component.WithTags("helmchart"),
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

	// Add helmchart annotations
	// Format: registry/repository (combining what was oci-registry and oci-repository)
	helmchartPath := fmt.Sprintf("%s/%s", registryHostname, repo)
	comp.SetAnnotation("giantswarm.io/helmcharts", helmchartPath)
	comp.SetAnnotation("giantswarm.io/helmchart-versions", chartVersion)

	// Add app version if available
	if appVersion != "" {
		comp.SetAnnotation("giantswarm.io/helmchart-app-versions", appVersion)
	}

	// Add audience and managed annotations
	comp.SetAnnotation(audienceBackstageAnnotation, audience)
	comp.SetAnnotation(managedBackstageAnnotation, strconv.FormatBool(managed))

	// Add icon URL if available
	if iconURL != "" {
		comp.SetAnnotation(iconBackstageAnnotation, iconURL)
	}

	// Add helmchart-deployable tag if the chart is deployable
	if componentutil.IsChartDeployable(chartType) {
		comp.AddTag("helmchart-deployable")
	}

	return comp, nil
}

// shouldIncludeChart checks if a chart should be included in the output based on
// the audience annotation in the config blob. Only charts with
// io.giantswarm.application.audience set to "all" are included.
func shouldIncludeChart(configMap map[string]interface{}) bool {
	if configMap == nil {
		return false
	}

	// Extract annotations from config blob
	annotations, ok := configMap["annotations"].(map[string]interface{})
	if !ok {
		return false
	}

	// Check for OCI annotation first, then legacy annotation
	var audienceValue string
	if val, ok := annotations[audienceOciAnnotation].(string); ok {
		audienceValue = val
	} else if val, ok := annotations[audienceLegacyChartAnnnotation].(string); ok {
		audienceValue = val
	} else {
		// No audience annotation found, exclude the chart
		return false
	}

	// Only include charts with audience set to "all"
	return audienceValue == audienceAll
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
