// Reads Giant Swarm installations info from the giantswarm/installations repository
package installations

import (
	"context"
	"log"
	"path/filepath"
	"slices"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v65/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// Names we don't consider actual installations
var reservedNames = []string{".github", "docs", "default"}

const (
	clusterFile = "cluster.yaml"
)

type Config struct {
	// Name of the GitHub organization owning our repository.
	GithubOrganization string

	// Name of the repository containing our repositories config.
	GithubRepositoryName string

	// Github personal access token (PTA) to use for client authentication.
	GithubAuthToken string
}

type Service struct {
	config       Config
	githubClient *github.Client
	ctx          context.Context
}

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

	s := &Service{
		ctx:          ctx,
		config:       c,
		githubClient: github.NewClient(tc),
	}

	return s, nil
}

// GetInstallations returns a slice of installations.
func (s *Service) GetInstallations() ([]*Installation, error) {
	_, directoryContent, _, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, s.config.GithubRepositoryName, "/", nil)
	if err != nil {
		return nil, err
	}

	paths := []string{}
	for _, item := range directoryContent {
		if item.GetType() == "dir" {
			name := item.GetName()
			if slices.Contains(reservedNames, name) {
				continue
			}
			paths = append(paths, name)
		}
	}

	ins := make([]*Installation, 0)

	for _, path := range paths {
		fullPath := filepath.Join(path, clusterFile)
		fileContent, _, _, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, s.config.GithubRepositoryName, fullPath, nil)
		if err != nil {
			return nil, err
		}

		if fileContent == nil {
			log.Printf("WARNING: file %s not found in repository %s/%s", fullPath, s.config.GithubOrganization, s.config.GithubRepositoryName)
			continue
		}

		content, err := fileContent.GetContent()
		if err != nil {
			log.Fatalf("INFO: error fetching content for file %s in repository %s/%s: %s", fullPath, s.config.GithubOrganization, s.config.GithubRepositoryName, err)
		}

		installation, err := parseInstallationInfo([]byte(content))
		if err != nil {
			log.Fatalf("WARNING: error parsing content for file %s: %s", fullPath, err)
		}

		ins = append(ins, installation)
	}

	return ins, nil
}

// GetInstallationFile returns the content of a single installation file.
func (s *Service) GetInstallationFile(name string) (string, error) {
	path := name + "/" + clusterFile
	fileContent, _, _, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, s.config.GithubRepositoryName, path, nil)
	if err != nil {
		return "", err
	}

	if fileContent == nil {
		return "", microerror.Maskf(fileNotFoundError, "file %s not found in repository %s", path, name)
	}

	return fileContent.GetContent()
}

func parseInstallationInfo(content []byte) (*Installation, error) {
	inst := &Installation{}
	err := yaml.Unmarshal(content, inst)
	if err != nil {
		return nil, err
	}

	// Copy region key to top position
	if inst.Region == "" && inst.Aws.Region != "" {
		inst.Region = inst.Aws.Region
	}

	return inst, nil
}
