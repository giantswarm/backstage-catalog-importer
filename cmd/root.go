// Package cmd contains CLI commands.
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/google/go-github/v53/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"

	"github.com/giantswarm/backstage-catalog-importer/pkg/catalog"
	"github.com/giantswarm/backstage-catalog-importer/pkg/repositories"
)

var rootCmd = &cobra.Command{
	Use:   "backstage-util",
	Short: "Giant Swarm tool to import data into backstage's catalog",
	Run:   runRoot,
}

const (
	// Name of the GitHub organization owning our teams and users.
	githubOrganization = "giantswarm"
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

	lists, err := repositories.GetLists(token)
	if err != nil {
		log.Fatal(err)
	}

	// Collect team names for Group entity creation.
	teamNamesMap := make(map[string]bool, 1)

	// Collect user names for User entity creation.
	userNamesMap := make(map[string]bool, 1)

	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	numComponents := 0
	numTeams := 0
	numUsers := 0

	// Iterate repository lists (per team) and create component entities.
	for _, list := range lists {
		teamNamesMap[list.OwnerTeamName] = true

		log.Printf("Processing %d repos of team %q\n", len(list.Repositories), list.OwnerTeamName)

		for _, repo := range list.Repositories {
			ent := createComponentEntity(repo, list.OwnerTeamName)
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

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Export teams
	teamNames := getMapKeys(teamNamesMap)
	for _, teamSlug := range teamNames {
		log.Printf("Group %s", teamSlug)
		team, _, err := client.Teams.GetTeamBySlug(ctx, githubOrganization, teamSlug)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		members, _, err := client.Teams.ListTeamMembersBySlug(ctx, githubOrganization, teamSlug, nil)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		var memberNames []string
		for _, u := range members {
			n := u.GetLogin()
			memberNames = append(memberNames, n)
			userNamesMap[n] = true
		}

		entity := createGroupEntity(teamSlug, team.GetName(), team.GetDescription(), *team.Parent.Name, memberNames, team.GetID())

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

	// Export users
	userNames := getMapKeys(userNamesMap)
	for _, userSlug := range userNames {
		log.Printf("User %s", userSlug)

		// load user data from Github
		user, _, err := client.Users.Get(ctx, userSlug)
		if err != nil {
			log.Fatalf("Error: %v", err)
		}

		entity := createUserEntity(userSlug, user.GetEmail(), user.GetName(), user.GetBio(), user.GetAvatarURL())

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

	log.Printf("Wrote %d components, %d teams, %d users", numComponents, numTeams, numUsers)
	log.Printf("Wrote output to %s", path)
}

func createComponentEntity(r repositories.Repo, team string) catalog.Entity {
	e := catalog.Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       catalog.EntityKindComponent,
		Metadata: catalog.EntityMetadata{
			Name:   r.Name,
			Labels: map[string]string{},
			Annotations: map[string]string{
				"github.com/project-slug":      fmt.Sprintf("giantswarm/%s", r.Name),
				"github.com/team-slug":         team,
				"backstage.io/source-location": fmt.Sprintf("url:https://github.com/giantswarm/%s", r.Name),
				"circleci.com/project-slug":    fmt.Sprintf("github/giantswarm/%s", r.Name),
				"quay.io/repository-slug":      fmt.Sprintf("giantswarm/%s", r.Name),
			},
			Tags: []string{},
		},
	}

	spec := catalog.ComponentSpec{
		Type:      "service",
		Lifecycle: "production",
		Owner:     team,
	}

	if r.Lifecycle != "production" && r.Lifecycle != "" {
		spec.Lifecycle = string(r.Lifecycle)
	}

	e.Spec = spec

	if r.Gen.Language != "" && r.Gen.Language != repositories.RepoLanguageGeneric {
		e.Metadata.Labels["giantswarm.io/language"] = string(r.Gen.Language)

		e.Metadata.Tags = append(e.Metadata.Tags, fmt.Sprintf("language:%s", r.Gen.Language))
	}

	for _, flavor := range r.Gen.Flavors {
		e.Metadata.Labels[fmt.Sprintf("giantswarm.io/flavor-%s", flavor)] = "true"

		e.Metadata.Tags = append(e.Metadata.Tags, fmt.Sprintf("flavor:%s", flavor))
	}

	return e
}

func createGroupEntity(name, displayName, description, parent string, members []string, id int64) catalog.Entity {
	e := catalog.Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       catalog.EntityKindGroup,
		Metadata: catalog.EntityMetadata{
			Name: name,
		},
	}
	spec := catalog.GroupSpec{
		Type:    "team",
		Members: members,
		Profile: catalog.GroupProfile{
			Picture: fmt.Sprintf("https://avatars.githubusercontent.com/t/%d?s=116&v=4", id),
		},
	}

	if description != "" {
		e.Metadata.Description = description
	}
	if displayName != "" {
		spec.Profile.DisplayName = displayName
	}
	if parent != "" {
		spec.Parent = parent
	}

	e.Spec = spec

	return e
}

func createUserEntity(name, email, displayName, description, avatarURL string) catalog.Entity {
	e := catalog.Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       catalog.EntityKindUser,
		Metadata: catalog.EntityMetadata{
			Name: name,
		},
	}

	spec := catalog.UserSpec{
		MemberOf: []string{},
		Profile: catalog.UserProfile{
			Email: email,
		},
	}

	if description != "" {
		e.Metadata.Description = description
	}
	if displayName != "" {
		spec.Profile.DisplayName = displayName
	}
	if avatarURL != "" {
		spec.Profile.Picture = avatarURL
	}

	e.Spec = spec

	return e
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
