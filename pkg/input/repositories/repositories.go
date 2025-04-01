// Package repositories provides types and tools to deal with
// Giant Swarm's repository configuration data maintained in
// https://github.com/giantswarm/github/tree/master/repositories
package repositories

import (
	"context"
	b64 "encoding/base64"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v70/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
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

// A service to access information Giant Swarm stores about
// GitHub repositories, as well as some additional details
// fetched from the GitHub API.
type Service struct {
	config       Config
	ctx          context.Context
	githubClient *github.Client

	// Cached information on certain repos
	githubRepoDetails map[string]GithubRepoDetails
	// Cached information on repo content
	githubRepoContentDetails map[string]GithubRepoContentDetails
	// Cached information on repo releases
	githubReleaseDetails map[string]GithubReleaseDetails
}

// New instantiates a new repositories service.
//
// This results in the fetching of basic data on all repositories owned by the organization,
// issuing one request per 100 repositories. Beware of rate limiting!
func New(c Config) (*Service, error) {
	if c.GithubOrganization == "" {
		return nil, microerror.Maskf(invalidConfigError, "no Github organization configured")
	}
	if c.GithubRepositoryName == "" {
		return nil, microerror.Maskf(invalidConfigError, "no Github repository name configured")
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
		config:                   c,
		ctx:                      ctx,
		githubClient:             client,
		githubRepoDetails:        make(map[string]GithubRepoDetails),
		githubRepoContentDetails: make(map[string]GithubRepoContentDetails),
		githubReleaseDetails:     make(map[string]GithubReleaseDetails),
	}

	if c.GithubAuthToken != "" {
		err := s.loadGithubRepoDetails()
		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Load main information for all repositories of the organization.
//
// This also loads info for archived repos, but it's still cheaper
// to do it this way than one by one, as we get up to 100 repos
// per request.
func (s *Service) loadGithubRepoDetails() error {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	repos := make(map[string]GithubRepoDetails)

	for {
		r, resp, err := s.githubClient.Repositories.ListByOrg(s.ctx, s.config.GithubOrganization, opts)
		if err != nil {
			return err
		}

		for _, repo := range r {
			if repo.GetArchived() {
				continue
			}

			name := repo.GetName()
			details := GithubRepoDetails{
				Name:          name,
				Description:   repo.GetDescription(),
				IsPrivate:     repo.GetPrivate(),
				DefaultBranch: repo.GetDefaultBranch(),
				MainLanguage:  strings.ToLower(repo.GetLanguage()),
			}

			repos[name] = details
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	s.githubRepoDetails = repos

	return nil
}

// Load details found in certain files in the repository.
func (s *Service) loadGithubRepoContentDetails(name string) error {
	details := GithubRepoContentDetails{}

	// Detect CircleCI
	_, _, resp, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, name, ".circleci/config.yml", nil)
	if err == nil {
		details.HasCircleCI = true
	} else if resp.StatusCode != http.StatusNotFound {
		// 404 is a "not found" error, which is expected. Everything else is not expected.
		return err
	}

	// Detect README
	_, _, resp, err = s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, name, "README.md", nil)
	if err == nil {
		details.HasReadme = true
	} else if resp.StatusCode != http.StatusNotFound {
		// 404 is a "not found" error, which is expected. Everything else is not expected.
		return err
	}

	// Detect helm folder
	_, directoryContent, resp, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, name, "helm", nil)
	if err == nil {
		if directoryContent != nil {
			details.HasHelmFolder = true
			details.NumHelmCharts = len(directoryContent)
			details.HelmChartNames = make([]string, len(directoryContent))
			for i, item := range directoryContent {
				details.HelmChartNames[i] = item.GetName()
			}
		}
	} else if resp.StatusCode != http.StatusNotFound {
		// 404 is a "not found" error, which is expected. Everything else is not expected.
		return err
	}

	s.githubRepoContentDetails[name] = details

	return nil
}

// Return the content of a source file in a repository as string.
func (s *Service) LoadGitHubFile(name string, path string) (string, error) {
	fileContent, _, _, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, name, path, nil)
	if err != nil {
		return "", err
	}

	if fileContent == nil {
		return "", microerror.Maskf(fileNotFoundError, "file %s not found in repository %s", path, name)
	}

	return fileContent.GetContent()
}

// Load tag and date of the projectz's latest release.
func (s *Service) loadGithubReleaseDetails(name string) error {
	latestRelease, response, err := s.githubClient.Repositories.GetLatestRelease(s.ctx, s.config.GithubOrganization, name)
	if err != nil && response.StatusCode != http.StatusNotFound {
		return err
	}

	created := time.Time{}
	tag := latestRelease.GetTagName()
	if tag != "" {
		created = latestRelease.CreatedAt.Time
	}

	details := GithubReleaseDetails{
		LatestReleaseTime: created,
		LatestReleaseTag:  tag,
	}

	s.githubReleaseDetails[name] = details

	return nil
}

// Loads a list of repository configurations from a local path.
// The file name is asserted in the format `<team_name>.yaml`, with all
// repositories mentioned in it belonging to the team of that name.
func (s *Service) loadList(path string) ([]Repo, error) {
	path = filepath.Clean(path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return s.loadListFromBytes(data)
}

func (s *Service) loadListFromBytes(data []byte) ([]Repo, error) {
	repos := []Repo{}
	err := yaml.Unmarshal(data, &repos)
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
		return ""
	}

	return s.githubRepoDetails[name].Description
}

// Returns the main language for the given repo. If not available,
// or an error occurs, returns an empty string.
func (s *Service) MustGetLanguage(name string) string {
	if _, ok := s.githubRepoDetails[name]; !ok {
		return ""
	}

	return s.githubRepoDetails[name].MainLanguage
}

// Returns the public/private info for the given repo.
func (s *Service) GetIsPrivate(name string) (bool, error) {
	if _, ok := s.githubRepoDetails[name]; !ok {
		return false, microerror.Maskf(repositoryNotFoundError, "repository %s not found", name)
	}

	return s.githubRepoDetails[name].IsPrivate, nil
}

// Returns the default branch name. Returns an empty string in case of error.
func (s *Service) MustGetDefaultBranch(name string) string {
	if _, ok := s.githubRepoDetails[name]; !ok {
		return ""
	}

	return s.githubRepoDetails[name].DefaultBranch
}

// Returns the creation time of the repository's most recent release.
// If no release was found, returns zero-value time.
func (s *Service) MustGetLatestReleaseTime(name string) (time.Time, error) {
	if _, ok := s.githubReleaseDetails[name]; !ok {
		err := s.loadGithubReleaseDetails(name)
		if err != nil {
			return time.Time{}, microerror.Mask(err)
		}
	}

	return s.githubReleaseDetails[name].LatestReleaseTime, nil
}

// Returns the tag of the repository's most recent release.
// If no release was found, returns empty string.
func (s *Service) MustGetLatestReleaseTag(name string) (string, error) {
	if _, ok := s.githubReleaseDetails[name]; !ok {
		err := s.loadGithubReleaseDetails(name)
		if err != nil {
			return "string", microerror.Mask(err)
		}
	}

	return s.githubReleaseDetails[name].LatestReleaseTag, nil
}

// Returns whether the repo has a CircleCI configuration.
func (s *Service) GetHasCircleCI(name string) (bool, error) {
	if _, ok := s.githubRepoContentDetails[name]; !ok {
		err := s.loadGithubRepoContentDetails(name)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	return s.githubRepoContentDetails[name].HasCircleCI, nil
}

// Returns whether the repo has a main README file.
func (s *Service) GetHasReadme(name string) (bool, error) {
	if _, ok := s.githubRepoContentDetails[name]; !ok {
		err := s.loadGithubRepoContentDetails(name)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	return s.githubRepoContentDetails[name].HasReadme, nil
}

// Returns whether the repo has a Helm chart.
func (s *Service) GetNumHelmCharts(name string) (int, error) {
	if _, ok := s.githubRepoContentDetails[name]; !ok {
		err := s.loadGithubRepoContentDetails(name)
		if err != nil {
			return 0, microerror.Mask(err)
		}
	}

	return s.githubRepoContentDetails[name].NumHelmCharts, nil
}

// Returns the name(s) of the repo's Helm chart(s).
func (s *Service) GetHelmChartNames(name string) ([]string, error) {
	if _, ok := s.githubRepoContentDetails[name]; !ok {
		err := s.loadGithubRepoContentDetails(name)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return s.githubRepoContentDetails[name].HelmChartNames, nil
}
