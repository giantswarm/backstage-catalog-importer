// Package cmd contains CLI commands.
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/cmd/charts"
	installations "github.com/giantswarm/backstage-catalog-importer/cmd/installations"
	users "github.com/giantswarm/backstage-catalog-importer/cmd/users"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/helmchart"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/repositories"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/teams"
	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/component"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/group"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/export"
	componentutil "github.com/giantswarm/backstage-catalog-importer/pkg/util/component"
)

var rootCmd = &cobra.Command{
	Use:   "backstage-catalog-importer",
	Short: "Giant Swarm tool to import data into backstage's catalog",
	Run:   runRoot,
}

const (
	// Name of the GitHub organization owning our teams and users.
	githubOrganization = "giantswarm"

	// Name of the repository holding our repository meta data.
	githubManagementRepository = "github"

	// Directory path within githubManagementRepository holding repo metadata YAML files.
	repositoriesPath = "repositories"
)

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", ".", "Output directory path")
	rootCmd.Flags().StringP("chart-repo-prefix", "", "charts/giantswarm", "Prefix for chart repositories in the OCI registries")
	rootCmd.Flags().StringP("public-oci-registry", "", "gsoci.azurecr.io", "Host name of the public OCI registry")
	rootCmd.Flags().StringP("private-oci-registry", "", "gsociprivate.azurecr.io", "Host name of the private OCI registry")

	rootCmd.AddCommand(charts.Command)
	rootCmd.AddCommand(installations.Command)
	rootCmd.AddCommand(users.Command)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runRoot(cmd *cobra.Command, args []string) {
	path, err := cmd.PersistentFlags().GetString("output")
	if err != nil {
		log.Fatal(err)
	}

	repoPrefix, err := cmd.Flags().GetString("chart-repo-prefix")
	if err != nil {
		log.Fatal(err)
	}
	// Remove leading and trailing slash if present
	repoPrefix = strings.TrimPrefix(repoPrefix, "/")
	repoPrefix = strings.TrimSuffix(repoPrefix, "/")

	publicOciRegistry, err := cmd.Flags().GetString("public-oci-registry")
	if err != nil {
		log.Fatal(err)
	}

	privateOciRegistry, err := cmd.Flags().GetString("private-oci-registry")
	if err != nil {
		log.Fatal(err)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("Please set environment variable GITHUB_TOKEN to a personal GitHub access token (PAT).")
	}

	repoService, err := repositories.New(repositories.Config{
		GithubOrganization:   githubOrganization,
		GithubRepositoryName: githubManagementRepository,
		GithubAuthToken:      token,
		DirectoryPath:        repositoriesPath,
	})
	if err != nil {
		log.Fatal(err)
	}

	teamsService, err := teams.New(teams.Config{
		GithubOrganization: githubOrganization,
		GithubAuthToken:    token,
	})
	if err != nil {
		log.Fatal(err)
	}

	lists, err := repoService.GetLists()
	if err != nil {
		log.Fatal(err)
	}

	groupExporter := export.New(export.Config{TargetPath: path + "/groups.yaml"})
	componentExporter := export.New(export.Config{TargetPath: path + "/components.yaml"})

	numComponents := 0
	numGroups := 0

	// Collect Go dependencies for later analysis
	dependencies := make(map[string][]string)
	repositoriesImported := make(map[string]bool)

	// Iterate repository lists (per team) and create component entities.
	for _, list := range lists {
		log.Printf("Processing %d repos of team %q\n", len(list.Repositories), list.OwnerTeamName)

		for _, repo := range list.Repositories {

			ociRegistry := publicOciRegistry
			isPrivate, err := repoService.GetIsPrivate(repo.Name)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			if isPrivate {
				ociRegistry = privateOciRegistry
			}

			hasReadme, err := repoService.GetHasReadme(repo.Name)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			// Fetch Helm chart info if available.
			var charts []*helmchart.Chart
			var hasDeployableChart bool
			{
				numCharts, err := repoService.GetNumHelmCharts(repo.Name)
				if err != nil {
					log.Fatalf("Error: %v", err)
				} else if numCharts > 0 {
					chartNames, _ := repoService.GetHelmChartNames(repo.Name)
					for _, chartName := range chartNames {
						log.Printf("DEBUG - %s - fetching info on helm chart %s\n", repo.Name, chartName)
						path := fmt.Sprintf("helm/%s/Chart.yaml", chartName)
						data, err := repoService.LoadGitHubFile(repo.Name, path)
						if err != nil {
							if !repositories.IsFileNotFoundError(err) {
								log.Printf("WARN - %s - error fetching helm chart %s: %v", repo.Name, chartName, err)
							}
						} else {
							chart, err := helmchart.LoadString(data)
							if err != nil {
								log.Printf("WARN - %s - error parsing helm chart %s: %v", repo.Name, chartName, err)
							} else {
								charts = append(charts, chart)
								if componentutil.IsChartDeployable(chart.Type) {
									hasDeployableChart = true
								}
							}
						}
					}
				}
			}

			deps := []string{}
			lang := repoService.MustGetLanguage(repo.Name)

			latestReleaseTime, err := repoService.MustGetLatestReleaseTime(repo.Name)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			latestReleaseTag, err := repoService.MustGetLatestReleaseTag(repo.Name)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			if lang == "go" {
				deps, err = repoService.GetDependencies(repo.Name)
				if err != nil {
					log.Printf("WARN - %s: error fetching dependencies: %v", repo.Name, err)
				}

				for _, d := range deps {
					dependencies[d] = append(dependencies[d], fmt.Sprintf("used in [%s](https://github.com/%s/%s) owned by @%s/%s", repo.Name, githubOrganization, repo.Name, githubOrganization, list.OwnerTeamName))
				}
			}

			// Prepare deployment names (default to <name> and <name>-app if not set).
			deploymentNames := repo.DeploymentNames
			if len(deploymentNames) == 0 {
				deploymentNames = componentutil.GenerateDeploymentNames(repo.Name)
			}

			description := repoService.MustGetDescription(repo.Name)
			defaultBranch := repoService.MustGetDefaultBranch(repo.Name)

			genLanguage := ""
			if repo.Gen.Language != "" && repo.Gen.Language != repositories.RepoLanguageGeneric {
				genLanguage = string(repo.Gen.Language)
			}

			genFlavors := make([]string, len(repo.Gen.Flavors))
			for i, flavor := range repo.Gen.Flavors {
				genFlavors[i] = string(flavor)
			}

			c, err := component.New(
				repo.Name,
				component.WithCircleCiSlug(fmt.Sprintf("github/%s/%s", githubOrganization, repo.Name)),
				component.WithDefaultBranch(defaultBranch),
				component.WithDependsOn(deps...),
				component.WithDeploymentNames(deploymentNames...),
				component.WithDescription(description),
				component.WithFlavors(genFlavors...),
				component.WithGithubProjectSlug(fmt.Sprintf("%s/%s", githubOrganization, repo.Name)),
				component.WithGithubTeamSlug(list.OwnerTeamName),
				component.WithHasReadme(hasReadme),
				component.WithHasReleases(latestReleaseTag != ""),
				component.WithHelmCharts(charts...),
				component.WithLanguage(genLanguage),
				component.WithLatestReleaseTag(latestReleaseTag),
				component.WithLatestReleaseTime(latestReleaseTime),
				component.WithLifecycle(string(repo.Lifecycle)),
				component.WithOwner(list.OwnerTeamName),
				component.WithPrivate(isPrivate),
				component.WithSystem(repo.System),
				component.WithType(repo.ComponentType),
				component.WithOciRegistry(ociRegistry),
				component.WithOciRepositoryPrefix(repoPrefix),
			)
			if err != nil {
				log.Fatalf("Could not create component: %s", err)
			}

			if hasDeployableChart {
				c.AddTag("helmchart-deployable")
			}

			if len(charts) > 0 {
				// Determine the chart's audience annotation, and if 'all', add tag.
				// (This ignores the rare case where a repo may have more than one chart with different audience annotations.)
				audience := ""
				for _, chart := range charts {
					if chart.Annotations != nil {
						if val, exists := chart.Annotations["io.giantswarm.application.audience"]; exists {
							audience = val
						}
					}
				}
				if audience == "all" {
					c.AddTag("helmchart-audience-all")
				}
			}

			// Grafana dashboard link for services.
			if repo.ComponentType == "service" {
				urlParts := []string{}
				for _, d := range deploymentNames {
					urlParts = append(urlParts, fmt.Sprintf("var-app=%s", d))
				}
				c.AddLink(bscatalog.EntityLink{
					URL:   fmt.Sprintf("https://giantswarm.grafana.net/d/eb617ba1-209a-4d57-9963-1af9a8ddc8d4/general-service-metrics?orgId=1&%s&from=now-24h&to=now", strings.Join(urlParts, "&")),
					Title: "General service metrics dashboard",
					Icon:  "dashboard",
					Type:  "grafana-dashboard",
				})
			}

			entity := c.ToEntity()
			numComponents++

			err = componentExporter.AddEntity(entity)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			repositoriesImported[repo.Name] = true
		}
	}

	// Export teams
	teams, err := teamsService.GetAll()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	log.Printf("Processing %d teams", len(teams))

	for _, team := range teams {
		members, err := teamsService.GetMembers(team.GetSlug())
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		var memberNames []string
		for _, u := range members {
			n := u.GetLogin()
			memberNames = append(memberNames, n)
		}

		parentTeamName := ""
		if team.GetParent() != nil {
			parentTeamName = team.GetParent().GetSlug()
		}

		group, err := group.New(team.GetSlug(),
			group.WithTitle(team.GetName()),
			group.WithDescription(team.GetDescription()),
			group.WithPictureURL(fmt.Sprintf("https://avatars.githubusercontent.com/t/%d?s=116&v=4", team.GetID())),
			group.WithMemberNames(memberNames...),
			group.WithParentName(parentTeamName),
			group.WithGrafanaDashboardSelector(fmt.Sprintf("tags @> 'owner:%s'", team.GetSlug())),
		)
		if err != nil {
			log.Fatalf("Error: could not create group -- %v", err)
		}

		numGroups++
		entity := group.ToEntity()
		err = groupExporter.AddEntity(entity)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}

	err = componentExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error writing components: %v", err)
	}
	err = groupExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error writing groups: %v", err)
	}

	// Filter Go dependencies to those that are not in the catalog.
	dependenciesNotCovered := make(map[string][]string)
	for name, info := range dependencies {
		ok := repositoriesImported[name]
		if !ok {
			dependenciesNotCovered[name] = info
		}
	}

	if len(dependenciesNotCovered) > 0 {
		fmt.Println("\nFound the following Go dependencies not covered in the catalog:")
		for name, info := range dependenciesNotCovered {
			fmt.Printf("\n- [ ] [%s](https://github.com/%s/%s)", name, githubOrganization, name)
			for _, infoItem := range info {
				fmt.Printf("\n   - %s", infoItem)
			}
		}

		fmt.Println("")
	}

	fmt.Printf("\n%d components written to file %s with size %d bytes", numComponents, componentExporter.TargetPath, componentExporter.Len())
	fmt.Printf("\n%d groups written to file %s with size %d bytes", numGroups, groupExporter.TargetPath, groupExporter.Len())
	fmt.Println("")
}
