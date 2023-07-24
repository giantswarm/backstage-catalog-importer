// Package repositories provides types and tools to deal with
// Giant Swarm's repository configuration data maintained in
// https://github.com/giantswarm/github/tree/master/repositories
package repositories

import (
	"context"
	b64 "encoding/base64"
	"os"
	"strings"

	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

const (
	// Name of the GitHub organization owning our repository.
	githubOrganization = "giantswarm"

	// Name of the repository containing our repositories config.
	githubRepositoryName = "github"

	// Path within the repository containing repository config YAML lists.
	directoryPath = "repositories"
)

type ListResult struct {
	OwnerTeamName string
	Repositories  []Repo
}

// LoadList loads a list of repository configurations from a local path.
// The file name is asserted in the format `<team_name>.yaml`, with all
// repositories mentioned in it belonging to the team of that name.
func LoadList(path string) ([]Repo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return LoadListFromBytes(data)
}

func LoadListFromBytes(data []byte) ([]Repo, error) {
	repos := []Repo{}
	err := yaml.UnmarshalStrict(data, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// GetLists is loading the list of repository YAML files from GitHub giantswarm/github.
// This requires a personal access token.
func GetLists(githubToken string) ([]ListResult, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	_, directoryContent, _, err := client.Repositories.GetContents(ctx, githubOrganization, githubRepositoryName, directoryPath, nil)
	if err != nil {
		return nil, err
	}

	result := []ListResult{}

	for _, item := range directoryContent {
		if !strings.HasSuffix(*item.Name, ".yaml") {
			continue
		}

		fileContent, _, _, err := client.Repositories.GetContents(ctx, githubOrganization, githubRepositoryName, *item.Path, nil)
		if err != nil {
			return nil, err
		}

		decodedContent, _ := b64.StdEncoding.DecodeString(*fileContent.Content)
		lists, err := LoadListFromBytes(decodedContent)
		if err != nil {
			return nil, err
		}

		result = append(result, ListResult{
			OwnerTeamName: strings.TrimSuffix(*item.Name, ".yaml"),
			Repositories:  lists,
		})
	}

	return result, nil
}
