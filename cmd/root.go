// Package cmd contains CLI commands.
package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/cmd/appcatalogs"
	installations "github.com/giantswarm/backstage-catalog-importer/cmd/installations"
	users "github.com/giantswarm/backstage-catalog-importer/cmd/users"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/helmchart"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/repositories"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/teams"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/group"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/export"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/legacy"
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

	rootCmd.AddCommand(appcatalogs.Command)
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
			isPrivate, err := repoService.GetIsPrivate(repo.Name)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			hasCircleCi, err := repoService.GetHasCircleCI(repo.Name)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			hasReadme, err := repoService.GetHasReadme(repo.Name)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			// Fetch Helm chart info if available.
			var charts []*helmchart.Chart
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
				log.Printf("DEBUG - found %d dependencies for repo %q", len(deps), repo.Name)

				for _, d := range deps {
					dependencies[d] = append(dependencies[d], fmt.Sprintf("used in [%s](https://github.com/%s/%s) owned by @%s/%s", repo.Name, githubOrganization, repo.Name, githubOrganization, list.OwnerTeamName))
				}
			}

			ent := legacy.CreateComponentEntity( //nolint:staticcheck
				repo,
				list.OwnerTeamName,
				repoService.MustGetDescription(repo.Name),
				repo.System,
				isPrivate,
				hasCircleCi,
				hasReadme,
				repoService.MustGetDefaultBranch(repo.Name),
				latestReleaseTime,
				latestReleaseTag,
				charts,
				deps)
			numComponents++

			err = componentExporter.AddEntity(&ent)
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
