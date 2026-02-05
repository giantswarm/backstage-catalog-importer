// Package githuburl provides functionality to fetch content from GitHub URLs.
package githuburl

import (
	"context"
	"strings"

	"github.com/giantswarm/microerror"
	"github.com/google/go-github/v82/github"
	"go.yaml.in/yaml/v3"
	"golang.org/x/oauth2"
)

// Config holds the service configuration.
type Config struct {
	// AuthToken is the GitHub authentication token.
	// If empty, unauthenticated requests will be made (lower rate limits).
	AuthToken string
}

// Service provides GitHub URL fetching functionality.
type Service struct {
	client *github.Client
	ctx    context.Context
}

// New creates a new GitHub URL fetching service.
func New(c Config) (*Service, error) {
	ctx := context.Background()

	var client *github.Client
	if c.AuthToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: c.AuthToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	return &Service{
		client: client,
		ctx:    ctx,
	}, nil
}

// FetchContent fetches the content from a GitHub URL.
// Supports both blob URLs and raw.githubusercontent.com URLs.
func (s *Service) FetchContent(url string) (string, error) {
	// Parse the URL to extract owner, repo, ref, and path
	owner, repo, ref, path, err := ParseGitHubURL(url)
	if err != nil {
		return "", microerror.Mask(err)
	}

	opts := &github.RepositoryContentGetOptions{}
	if ref != "" {
		opts.Ref = ref
	}

	fileContent, _, _, err := s.client.Repositories.GetContents(
		s.ctx,
		owner,
		repo,
		path,
		opts,
	)
	if err != nil {
		return "", microerror.Maskf(fetchError, "failed to fetch content from %s: %v", url, err)
	}

	if fileContent == nil {
		return "", microerror.Maskf(fetchError, "URL points to a directory, not a file: %s", url)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", microerror.Maskf(fetchError, "failed to decode content from %s: %v", url, err)
	}

	return content, nil
}

// ParseGitHubURL extracts owner, repo, ref, and path from a GitHub URL.
// Supports formats:
//   - https://github.com/owner/repo/blob/ref/path/to/file.yaml
//   - https://raw.githubusercontent.com/owner/repo/ref/path/to/file.yaml
func ParseGitHubURL(url string) (owner, repo, ref, path string, err error) {
	// Normalize the URL
	url = strings.TrimSpace(url)
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Handle raw.githubusercontent.com format
	if remainder, found := strings.CutPrefix(url, "raw.githubusercontent.com/"); found {
		parts := strings.SplitN(remainder, "/", 4)
		if len(parts) < 4 {
			return "", "", "", "", microerror.Maskf(invalidURLError, "invalid raw GitHub URL format: expected owner/repo/ref/path")
		}
		if parts[0] == "" || parts[1] == "" || parts[3] == "" {
			return "", "", "", "", microerror.Maskf(invalidURLError, "invalid GitHub URL: owner, repo, and path must not be empty")
		}
		return parts[0], parts[1], parts[2], parts[3], nil
	}

	// Handle github.com blob format
	if remainder, found := strings.CutPrefix(url, "github.com/"); found {
		parts := strings.SplitN(remainder, "/", 5)
		if len(parts) < 5 {
			return "", "", "", "", microerror.Maskf(invalidURLError, "invalid GitHub blob URL format: expected owner/repo/blob/ref/path")
		}
		if parts[2] != "blob" {
			return "", "", "", "", microerror.Maskf(invalidURLError, "invalid GitHub URL: expected 'blob' in path, got '%s'", parts[2])
		}
		if parts[0] == "" || parts[1] == "" || parts[4] == "" {
			return "", "", "", "", microerror.Maskf(invalidURLError, "invalid GitHub URL: owner, repo, and path must not be empty")
		}
		return parts[0], parts[1], parts[3], parts[4], nil
	}

	return "", "", "", "", microerror.Maskf(invalidURLError, "unsupported URL format: must be github.com or raw.githubusercontent.com")
}

// CRDMetadata contains metadata extracted from a CRD YAML.
type CRDMetadata struct {
	// Name is the full CRD name (e.g., "apps.application.giantswarm.io").
	Name string

	// Kind is the CRD kind (e.g., "App").
	Kind string

	// Group is the API group (e.g., "application.giantswarm.io").
	Group string

	// Description is extracted from the CRD schema if available.
	Description string
}

// ParseCRDMetadata extracts metadata from CRD YAML content.
func ParseCRDMetadata(content string) (*CRDMetadata, error) {
	// Use a simple struct to extract just what we need
	var crd struct {
		APIVersion string `yaml:"apiVersion"`
		Kind       string `yaml:"kind"`
		Metadata   struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
		Spec struct {
			Group string `yaml:"group"`
			Names struct {
				Kind     string `yaml:"kind"`
				Plural   string `yaml:"plural"`
				Singular string `yaml:"singular"`
			} `yaml:"names"`
			Versions []struct {
				Name   string `yaml:"name"`
				Schema struct {
					OpenAPIV3Schema struct {
						Description string `yaml:"description"`
					} `yaml:"openAPIV3Schema"`
				} `yaml:"schema"`
			} `yaml:"versions"`
		} `yaml:"spec"`
	}

	if err := yaml.Unmarshal([]byte(content), &crd); err != nil {
		return nil, microerror.Maskf(parseError, "failed to parse CRD YAML: %v", err)
	}

	// Validate it's a CRD
	if crd.Kind != "CustomResourceDefinition" {
		return nil, microerror.Maskf(parseError, "expected kind CustomResourceDefinition, got %s", crd.Kind)
	}

	if crd.Metadata.Name == "" {
		return nil, microerror.Maskf(parseError, "CRD is missing metadata.name")
	}

	if crd.Spec.Names.Kind == "" {
		return nil, microerror.Maskf(parseError, "CRD is missing spec.names.kind")
	}

	// Extract description from the first version's schema if available
	description := ""
	if len(crd.Spec.Versions) > 0 {
		description = crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Description
	}

	return &CRDMetadata{
		Name:        crd.Metadata.Name,
		Kind:        crd.Spec.Names.Kind,
		Group:       crd.Spec.Group,
		Description: description,
	}, nil
}
