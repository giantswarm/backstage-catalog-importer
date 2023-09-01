// Package cmd contains CLI commands.
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/google/go-github/v54/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

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
	rootCmd.PersistentFlags().StringP("format", "f", "raw", "Output format, 'raw' or 'configmap'.")
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

	format, err := cmd.PersistentFlags().GetString("format")
	if err != nil {
		log.Fatal(err)
	}

	if format != "raw" && format != "configmap" {
		log.Fatal("Invalid --format value. Please use 'raw' or 'configmap'.")
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

	var size int

	if format == "raw" {
		size = f.Len()
		_, err = file.WriteString(f.String())
		if err != nil {
			log.Fatal(err)
		}
	} else {
		cm := corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "catalog",
				Namespace: "backstage",
			},
			Data: map[string]string{
				"catalog.yaml": f.String(),
			},
		}

		serializer := json.NewYAMLSerializer(json.DefaultMetaFactory, nil, nil)
		var buf bytes.Buffer
		err := serializer.Encode(&cm, &buf)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		size = buf.Len()
		_, err = file.WriteString(buf.String())
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Wrote %d components, %d groups, %d users", numComponents, numTeams, numUsers)
	if format == "configmap" {
		log.Printf("Wrote ConfigMap to %s with size %d bytes", path, size)
	} else {
		log.Printf("Wrote YAML output to %s with %d bytes", path, size)
	}
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
