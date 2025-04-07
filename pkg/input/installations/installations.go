// Reads Giant Swarm installations info from the giantswarm/installations repository
package installations

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"slices"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v71/github"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

// Names we don't consider actual installations
var reservedNames = []string{
	".github",
	"docs",
	"default",
}

const (
	// Relative path of the file that proivides access instructions in Markdown format
	accessFile = "docs/access.md"

	// Name of the file that provides basic structured data on the installation
	clusterFile = "cluster.yaml"

	// Name of the file that provides the custom CA certificate for the installation (optional)
	caFile = "ca.pem"
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
	// Get installations list
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

	// Get default branch name from repo
	repo, _, err := s.githubClient.Repositories.Get(s.ctx, s.config.GithubOrganization, s.config.GithubRepositoryName)
	if err != nil {
		return nil, err
	}

	defaultBranch := repo.GetDefaultBranch()

	ins := make([]*Installation, 0)

	for _, path := range paths {
		// Load cluster file
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

		// Check if CA file exists
		caFilePath := filepath.Join(path, caFile)
		_, _, response, err := s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, s.config.GithubRepositoryName, caFilePath, nil)
		if err != nil {
			if response.StatusCode != http.StatusNotFound {
				return nil, err
			}
		}
		if response != nil && response.StatusCode == http.StatusOK {
			installation.CustomCA = fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s", s.config.GithubOrganization, s.config.GithubRepositoryName, defaultBranch, caFilePath)
		}

		// Check if access file exists
		accessFilePath := filepath.Join(path, accessFile)
		fileContent, _, response, err = s.githubClient.Repositories.GetContents(s.ctx, s.config.GithubOrganization, s.config.GithubRepositoryName, accessFilePath, nil)
		if err != nil {
			if response.StatusCode != http.StatusNotFound {
				return nil, err
			}
		}
		if response != nil && response.StatusCode == http.StatusOK {
			installation.AccessMarkdown, err = fileContent.GetContent()
			if err != nil {
				log.Fatalf("INFO: error fetching content for file %s in repository %s/%s: %s", fullPath, s.config.GithubOrganization, s.config.GithubRepositoryName, err)
			}
		}

		ins = append(ins, installation)
	}

	return ins, nil
}

func parseInstallationInfo(content []byte) (*Installation, error) {
	inst := &Installation{}
	err := yaml.Unmarshal(content, inst)
	if err != nil {
		return nil, err
	}

	// Copy AWS region key to top position
	if inst.Region == "" && inst.Aws != nil && inst.Aws.Region != "" {
		inst.Region = inst.Aws.Region
	}

	return inst, nil
}
