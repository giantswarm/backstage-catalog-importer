// Package teams provides means to export GitHub teams to Backstage group entities.
package teams

import (
	"context"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v62/github"
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

// New instantiates a new teams service.
func New(c Config) (*Service, error) {
	if c.GithubOrganization == "" {
		return nil, microerror.Maskf(invalidConfigError, "no Github organization configured")
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
		config:       c,
		ctx:          ctx,
		githubClient: client,
	}

	return s, nil
}

// Return all teams in the Github organization.
func (s *Service) GetAll() ([]*github.Team, error) {
	opts := &github.ListOptions{}
	teams := []*github.Team{}
	for {
		t, resp, err := s.githubClient.Teams.ListTeams(s.ctx, s.config.GithubOrganization, opts)
		if err != nil {
			return nil, err
		}

		teams = append(teams, t...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	return teams, nil
}

// Return one teams by its slug.
func (s *Service) GetBySlug(slug string) (*github.Team, error) {
	t, _, err := s.githubClient.Teams.GetTeamBySlug(s.ctx, s.config.GithubOrganization, slug)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Return member users for a team
func (s *Service) GetMembers(teamSlug string) ([]*github.User, error) {
	opts := &github.TeamListTeamMembersOptions{}
	members := []*github.User{}
	for {
		m, resp, err := s.githubClient.Teams.ListTeamMembersBySlug(s.ctx, s.config.GithubOrganization, teamSlug, opts)
		if err != nil {
			return nil, err
		}

		members = append(members, m...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return members, nil
}
