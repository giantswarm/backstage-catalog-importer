// Functions to fetch data about / from an individual Github repository.
package githubrepo

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v66/github"
	"golang.org/x/oauth2"
)

type Config struct {
	// Name of the GitHub organization owning our repository.
	GithubOrganization string

	// Github personal access token (PTA) to use for client authentication.
	GithubAuthToken string
}

type Service struct {
	config       Config
	ctx          context.Context
	githubClient *github.Client
}

// All the details of a GitHub repository that we care about.
type RepoDetails struct {
	Name          string
	Description   string
	IsPrivate     bool
	DefaultBranch string
	MainLanguage  string
}

// Instantiates a new service.
func New(c Config) (*Service, error) {
	if c.GithubOrganization == "" {
		return nil, microerror.Maskf(invalidConfigError, "no Github organization configured")
	}
	if c.GithubAuthToken == "" {
		log.Println("WARNING: No Github token given (env variable GITHUB_TOKEN not set)")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.GithubAuthToken},
	)

	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	s := &Service{
		config:       c,
		ctx:          ctx,
		githubClient: client,
	}

	return s, nil
}

// Load details for a specific repository.
func (s *Service) GetDetails(name string) (*RepoDetails, error) {
	repo, resp, err := s.githubClient.Repositories.Get(s.ctx, s.config.GithubOrganization, name)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, microerror.Maskf(repositoryNotFoundError, "repository %s not found", name)
	}

	details := &RepoDetails{
		Name:          repo.GetName(),
		Description:   repo.GetDescription(),
		IsPrivate:     repo.GetPrivate(),
		DefaultBranch: repo.GetDefaultBranch(),
		MainLanguage:  strings.ToLower(repo.GetLanguage()),
	}

	return details, nil
}
