// Package repositories provides types and tools to deal with
// Giant Swarm's repository configuration data maintained in
// https://github.com/giantswarm/github/tree/master/repositories
package repositories

import (
	"context"
	b64 "encoding/base64"
	"os"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v53/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	// Name of the GitHub organization owning our repository.
	GithubOrganization string

	// Name of the repository containing our repositories config.
	GithubRepositoryName string

	// Path within the repository containing repository config YAML lists.
	// An empty string indicates the root directory.
	DirectoryPath string
}

type ListResult struct {
	OwnerTeamName string
	Repositories  []Repo
}

type Service struct {
	config Config
}

func New(c Config) (*Service, error) {
	if c.GithubOrganization == "" {
		return nil, microerror.Maskf(invalidConfigError, "no Github organization configured")
	}
	if c.GithubRepositoryName == "" {
		return nil, microerror.Maskf(invalidConfigError, "no Github repository name configured")
	}

	s := &Service{
		config: c,
	}

	return s, nil
}

// LoadList loads a list of repository configurations from a local path.
// The file name is asserted in the format `<team_name>.yaml`, with all
// repositories mentioned in it belonging to the team of that name.
func (s *Service) LoadList(path string) ([]Repo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return s.LoadListFromBytes(data)
}

func (s *Service) LoadListFromBytes(data []byte) ([]Repo, error) {
	repos := []Repo{}
	err := yaml.UnmarshalStrict(data, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// GetLists is loading the list of repository YAML files from GitHub giantswarm/github.
// This requires a personal access token.
func (s *Service) GetLists(githubToken string) ([]ListResult, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	_, directoryContent, _, err := client.Repositories.GetContents(ctx, s.config.GithubOrganization, s.config.GithubRepositoryName, s.config.DirectoryPath, nil)
	if err != nil {
		return nil, err
	}

	result := []ListResult{}

	for _, item := range directoryContent {
		if !strings.HasSuffix(*item.Name, ".yaml") {
			continue
		}

		fileContent, _, _, err := client.Repositories.GetContents(ctx, s.config.GithubOrganization, s.config.GithubRepositoryName, *item.Path, nil)
		if err != nil {
			return nil, err
		}

		decodedContent, _ := b64.StdEncoding.DecodeString(*fileContent.Content)
		lists, err := s.LoadListFromBytes(decodedContent)
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
