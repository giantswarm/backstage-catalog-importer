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

	// Github personal access token (PTA) to use for client authentication.
	GithubAuthToken string

	// Path within the repository containing repository config YAML lists.
	// An empty string indicates the root directory.
	DirectoryPath string
}

type ListResult struct {
	OwnerTeamName string
	Repositories  []Repo
}

// GithubRepo is a sparce struct for just the GitHub repository info we need.
type GithubRepoDetails struct {
	Name        string
	Description string
	IsPrivate   bool
	HasCircleCI bool
	HasReadme   bool
}

type Service struct {
	config       Config
	ctx          context.Context
	githubClient *github.Client

	// Cached information on certain repos
	githubRepoDetails map[string]GithubRepoDetails
}

// New instantiates a new repositories service.
func New(c Config) (*Service, error) {
	if c.GithubOrganization == "" {
		return nil, microerror.Maskf(invalidConfigError, "no Github organization configured")
	}
	if c.GithubRepositoryName == "" {
		return nil, microerror.Maskf(invalidConfigError, "no Github repository name configured")
	}
	if c.GithubAuthToken == "" {
		return nil, microerror.Maskf(invalidConfigError, "no Github token given")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.GithubAuthToken},
	)

	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	s := &Service{
		config:            c,
		ctx:               ctx,
		githubClient:      client,
		githubRepoDetails: make(map[string]GithubRepoDetails),
	}

	return s, nil
}

// Load information for a specific repo from the Github API.
func (s *Service) loadGithubRepoDetails(name string) error {
	repo, _, err := s.githubClient.Repositories.Get(s.ctx, s.config.GithubOrganization, name)
	if err != nil {
		return err
	}

	details := GithubRepoDetails{
		Name:        name,
		Description: repo.GetDescription(),
		IsPrivate:   repo.GetPrivate(),
	}

	_, _, resp2, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, name, ".circleci/config.yml", nil)
	if err == nil {
		details.HasCircleCI = true
	} else if resp2.StatusCode != 404 {
		// 404 is a "not found" error, which is expected. Everything else is not expected.
		return err
	}

	_, _, resp3, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, name, "README.md", nil)
	if err == nil {
		details.HasReadme = true
	} else if resp3.StatusCode != 404 {
		// 404 is a "not found" error, which is expected. Everything else is not expected.
		return err
	}

	s.githubRepoDetails[name] = details

	return nil
}

// Loads a list of repository configurations from a local path.
// The file name is asserted in the format `<team_name>.yaml`, with all
// repositories mentioned in it belonging to the team of that name.
func (s *Service) loadList(path string) ([]Repo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return s.loadListFromBytes(data)
}

func (s *Service) loadListFromBytes(data []byte) ([]Repo, error) {
	repos := []Repo{}
	err := yaml.UnmarshalStrict(data, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

// GetLists loads the lists of repository YAML files from GitHub giantswarm/github.
func (s *Service) GetLists() ([]ListResult, error) {
	// Get repositories directory content.
	_, directoryContent, _, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, s.config.GithubRepositoryName, s.config.DirectoryPath, nil)
	if err != nil {
		return nil, err
	}

	result := []ListResult{}

	for _, item := range directoryContent {
		if !strings.HasSuffix(*item.Name, ".yaml") {
			continue
		}

		// Get individual team repositories file.
		fileContent, _, _, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, s.config.GithubRepositoryName, *item.Path, nil)
		if err != nil {
			return nil, err
		}

		decodedContent, _ := b64.StdEncoding.DecodeString(*fileContent.Content)
		lists, err := s.loadListFromBytes(decodedContent)
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

// Returns the description for the given repo. If not available,
// or an error occurs, returns an empty string.
func (s *Service) MustGetDescription(name string) string {
	if _, ok := s.githubRepoDetails[name]; !ok {
		err := s.loadGithubRepoDetails(name)
		if err != nil {
			return ""
		}
	}

	return s.githubRepoDetails[name].Description
}

// Returns the public/private info for the given repo.
func (s *Service) GetIsPrivate(name string) (bool, error) {
	if _, ok := s.githubRepoDetails[name]; !ok {
		err := s.loadGithubRepoDetails(name)
		if err != nil {
			return false, err
		}
	}

	return s.githubRepoDetails[name].IsPrivate, nil
}

// Returns whether the repo has a CircleCI configuration.
func (s *Service) GetHasCircleCI(name string) (bool, error) {
	if _, ok := s.githubRepoDetails[name]; !ok {
		err := s.loadGithubRepoDetails(name)
		if err != nil {
			return false, err
		}
	}

	return s.githubRepoDetails[name].HasCircleCI, nil
}

// Returns whether the repo has a main README file.
func (s *Service) GetHasReadme(name string) (bool, error) {
	if _, ok := s.githubRepoDetails[name]; !ok {
		err := s.loadGithubRepoDetails(name)
		if err != nil {
			return false, err
		}
	}

	return s.githubRepoDetails[name].HasReadme, nil
}
