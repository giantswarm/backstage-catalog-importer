// Package cmd contains CLI commands.
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/google/go-github/v55/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"

	"github.com/giantswarm/backstage-catalog-importer/pkg/catalog"
	"github.com/giantswarm/backstage-catalog-importer/pkg/repositories"
	"github.com/giantswarm/backstage-catalog-importer/pkg/teams"
)

var rootCmd = &cobra.Command{
	Use:   "backstage-util",
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
	rootCmd.PersistentFlags().StringP("output", "o", "output.yaml", "Output file path")
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

	// Output buffer
	var f bytes.Buffer

	numComponents := 0
	numTeams := 0
	numUsers := 0

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

			deps := []string{}
			lang := repoService.MustGetLanguage(repo.Name)

			if lang == "go" {
				deps, err = repoService.GetDependencies(repo.Name)
				if err != nil {
					log.Printf("WARN - %s: error fetching dependencies: %v", repo.Name, err)
				}

				for _, d := range deps {
					dependencies[d] = append(dependencies[d], fmt.Sprintf("used in [%s](https://github.com/%s/%s) owned by @%s/%s", repo.Name, githubOrganization, repo.Name, githubOrganization, list.OwnerTeamName))
				}
			}

			ent := catalog.CreateComponentEntity(
				repo,
				list.OwnerTeamName,
				repoService.MustGetDescription(repo.Name),
				repo.System,
				isPrivate,
				hasCircleCi,
				hasReadme,
				repoService.MustGetDefaultBranch(repo.Name),
				deps)
			numComponents++

			d, err := yaml.Marshal(&ent)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			_, err = f.WriteString("---\n")
			if err != nil {
				log.Fatalf("Error: %v", err)
			}
			_, err = f.Write(d)
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			repositoriesImported[repo.Name] = true
		}
	}

	// Collect user names for User entity creation.
	userNamesMap := make(map[string]bool, 1)

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
			userNamesMap[n] = true
		}

		parentTeamName := ""
		if team.GetParent() != nil {
			parentTeamName = team.GetParent().GetSlug()
		}

		entity := catalog.CreateGroupEntity(
			team.GetSlug(),
			team.GetName(),
			team.GetDescription(),
			parentTeamName,
			memberNames,
			team.GetID())

		numTeams++

		d, err := yaml.Marshal(&entity)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		_, err = f.WriteString("---\n")
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		_, err = f.Write(d)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Export users
	userNames := getMapKeys(userNamesMap)
	log.Printf("Processing %d users", len(userNames))

	for _, userSlug := range userNames {

		// load user data from Github
		user, _, err := client.Users.Get(ctx, userSlug)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		entity := catalog.CreateUserEntity(userSlug, user.GetEmail(), user.GetName(), user.GetBio(), user.GetAvatarURL())

		d, err := yaml.Marshal(&entity)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		_, err = f.WriteString("---\n")
		if err != nil {
			log.Fatalf("Error: %v", err)
		}
		_, err = f.Write(d)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		numUsers++
	}

	file, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = file.WriteString("#\n# This file was automatically generated, PLEASE DO NOT MODIFY IT BY HAND.\n#\n\n")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	size := f.Len()
	_, err = file.WriteString(f.String())
	if err != nil {
		log.Fatal(err)
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

	fmt.Printf("\nWrote %d components, %d groups, %d users", numComponents, numTeams, numUsers)
	fmt.Printf("\nWrote YAML output to %s with %d bytes\n", path, size)
}

// Returns a sorted slice of keys.
func getMapKeys(m map[string]bool) []string {
	keys := []string{}
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
