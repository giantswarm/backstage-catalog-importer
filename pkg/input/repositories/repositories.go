// Package repositories provides types and tools to deal with
// Giant Swarm's repository configuration data maintained in
// https://github.com/giantswarm/github/tree/master/repositories
package repositories

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v91/github"
	"go.yaml.in/yaml/v3"

	"github.com/giantswarm/backstage-catalog-importer/pkg/httpclient"
)

// valuesSchemaFileName is the values schema a deployable chart is expected to
// ship. app-build-suite's HasValuesSchema check looks for exactly this name.
const valuesSchemaFileName = "values.schema.json"

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

	ctx := context.Background()
	client, err := httpclient.NewGitHubClient(c.GithubAuthToken)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	s := &Service{
		config:                   c,
		ctx:                      ctx,
		githubClient:             client,
		githubRepoDetails:        make(map[string]GithubRepoDetails),
		githubRepoContentDetails: make(map[string]GithubRepoContentDetails),
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
	circleciFileContent, _, resp, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, name, ".circleci/config.yml", nil)
	if err == nil {
		details.HasCircleCI = true

		// Check if push-to-registries uses force-public: true
		if circleciFileContent != nil {
			content, contentErr := circleciFileContent.GetContent()
			if contentErr == nil {
				details.ForcePublicRegistry = circleciConfigHasForcePublic(content)
				if details.ForcePublicRegistry {
					log.Printf("DEBUG - %s - CircleCI config has force-public: true in push-to-registries\n", name)
				}
			}
		}
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
			details.HelmChartNames = make([]string, 0, len(directoryContent))
			details.HasValuesSchema = make(map[string]bool, len(directoryContent))
			for _, item := range directoryContent {
				// Only sub-directories are charts. A stray helm/README.md is
				// not one, and listing it would spend a request that returns
				// file content instead of a listing - and answers 403 rather
				// than 404 once the file passes 1 MB.
				if item.GetType() != "dir" {
					continue
				}

				chartName := item.GetName()
				details.HelmChartNames = append(details.HelmChartNames, chartName)

				// Detect values.schema.json per chart. One listing per chart
				// rather than one lookup per file, so Chart.yaml and anything
				// else we come to need is already in hand.
				_, chartContent, chartResp, chartErr := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, name, fmt.Sprintf("helm/%s", chartName), nil)
				if chartErr == nil {
					for _, chartItem := range chartContent {
						if chartItem.GetName() == valuesSchemaFileName {
							details.HasValuesSchema[chartName] = true

							break
						}
					}
					if _, seen := details.HasValuesSchema[chartName]; !seen {
						details.HasValuesSchema[chartName] = false
					}
				} else if chartResp != nil && chartResp.StatusCode != http.StatusNotFound {
					// Anything but "not found" means we do not know, so the
					// chart stays absent from the map rather than being
					// recorded as having no schema. It must not abort the
					// load: everything else about the repo is still valid, and
					// a single transient 403 or 502 among the per-chart
					// listings would otherwise kill the whole import, since
					// GetNumHelmCharts is fatal on error.
					log.Printf("WARN - %s - could not list helm/%s, schema presence unknown: %v\n", name, chartName, chartErr)
				}
			}
			details.NumHelmCharts = len(details.HelmChartNames)
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

// Returns whether the repo's CircleCI config uses force-public in push-to-registries,
// meaning charts/images go to the public registry despite the repo being private.
func (s *Service) GetForcePublicRegistry(name string) (bool, error) {
	if _, ok := s.githubRepoContentDetails[name]; !ok {
		err := s.loadGithubRepoContentDetails(name)
		if err != nil {
			return false, microerror.Mask(err)
		}
	}

	return s.githubRepoContentDetails[name].ForcePublicRegistry, nil
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

// Returns, per chart name, whether the repo carries
// helm/<chart>/values.schema.json. A chart absent from the returned map was not
// determined and must be treated as unknown, not as missing.
func (s *Service) GetHasValuesSchema(name string) (map[string]bool, error) {
	if _, ok := s.githubRepoContentDetails[name]; !ok {
		err := s.loadGithubRepoContentDetails(name)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	return s.githubRepoContentDetails[name].HasValuesSchema, nil
}

// circleciConfigHasForcePublic parses a CircleCI config YAML and checks whether
// any push-to-registries job in the workflows section has force-public: true.
func circleciConfigHasForcePublic(configYAML string) bool {
	var config struct {
		Workflows map[string]struct {
			Jobs []map[string]any `yaml:"jobs"`
		} `yaml:"workflows"`
	}

	if err := yaml.Unmarshal([]byte(configYAML), &config); err != nil {
		return false
	}

	for _, workflow := range config.Workflows {
		for _, job := range workflow.Jobs {
			for jobName, jobConfig := range job {
				// Match job names like "architect/push-to-registries" or "push-to-registries"
				if !strings.HasSuffix(jobName, "push-to-registries") {
					continue
				}
				params, ok := jobConfig.(map[string]any)
				if !ok {
					continue
				}
				if forcePublic, exists := params["force-public"]; exists {
					if val, ok := forcePublic.(bool); ok && val {
						return true
					}
				}
			}
		}
	}

	return false
}
