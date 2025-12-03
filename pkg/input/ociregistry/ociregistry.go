// Package ociregistry provides a means to read repositories from an OCI registry.
package ociregistry

import (
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/giantswarm/microerror"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

type Registry struct {
	Client *remote.Client

	registry *remote.Registry
}

type Config struct {
	Hostname string
}

func NewRegistry(ctx context.Context, config Config) (*Registry, error) {
	if config.Hostname == "" {
		return nil, microerror.Maskf(invalidConfigError, "hostname is required")
	}

	// Create registry client
	reg, err := remote.NewRegistry(config.Hostname)
	if err != nil {
		return nil, microerror.Maskf(couldNotCreateRegistryClientError, "error creating registry client: %v", err)
	}

	// Configure for anonymous access
	reg.Client = &auth.Client{
		Client: nil, // Use default HTTP client
		Cache:  auth.NewCache(),
	}
	reg.PlainHTTP = false // Use HTTPS

	return &Registry{
		Client:   &reg.Client,
		registry: reg,
	}, nil
}

// ListRepositories retrieves all repository names
// starting with prefix from the registry
func (r *Registry) ListRepositories(ctx context.Context, prefix string) ([]string, error) {
	var repos []string

	err := r.registry.Repositories(ctx, "", func(repoNames []string) error {
		for _, repoName := range repoNames {
			if strings.HasPrefix(repoName, prefix) {
				repos = append(repos, repoName)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return repos, nil
}

// ListRepositoryTags retrieves all tags for a given repository
func (r *Registry) ListRepositoryTags(ctx context.Context, repository string) ([]string, error) {
	var tags []string

	repo, err := r.registry.Repository(ctx, repository)
	if err != nil {
		return nil, microerror.Maskf(couldNotGetRepositoryError, "error getting repository: %v", err)
	}

	err = repo.Tags(ctx, "", func(tagNames []string) error {
		tags = append(tags, tagNames...)
		return nil
	})
	if err != nil {
		return nil, microerror.Maskf(couldNotGetRepositoryTagsError, "error getting repository tags: %v", err)
	}

	return tags, nil
}

// GetRepositoryManifest retrieves the manifest config for a given repository and tag
func (r *Registry) GetRepositoryManifest(ctx context.Context, repository, tag string) (map[string]interface{}, error) {
	repo, err := r.registry.Repository(ctx, repository)
	if err != nil {
		return nil, microerror.Maskf(couldNotGetRepositoryError, "error getting repository: %v", err)
	}

	// Resolve the tag to get descriptor with metadata
	descriptor, err := repo.Resolve(ctx, tag)
	if err != nil {
		return nil, microerror.Maskf(
			couldNotResolveTagError,
			"error resolving tag: %v",
			err,
		)
	}

	manifestConfig, err := fetchManifestConfig(ctx, repo, descriptor)
	if err != nil {
		return nil, microerror.Maskf(couldNotGetRepositoryManifestError, "error getting repository manifest: %v", err)
	}

	return manifestConfig, nil
}

// fetchManifestConfig fetches the config blob from a manifest and returns it as pretty JSON and parsed map
func fetchManifestConfig(ctx context.Context, repo registry.Repository, manifestDescriptor v1.Descriptor) (map[string]interface{}, error) {
	// Fetch the manifest
	manifestReader, err := repo.Fetch(ctx, manifestDescriptor)
	if err != nil {
		return nil, microerror.Maskf(couldNotGetRepositoryManifestError, "error fetching manifest: %v", err)
	}
	defer manifestReader.Close()

	// Read the manifest content
	manifestBytes, err := io.ReadAll(manifestReader)
	if err != nil {
		return nil, microerror.Maskf(couldNotReadManifestError, "error reading manifest: %v", err)
	}

	// Parse as OCI manifest
	var manifest v1.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		// Not a manifest (could be an index), skip config
		return nil, microerror.Maskf(couldNotUnmarshalManifestError, "error unmarshalling manifest: %v", err)
	}

	// Fetch the config blob using the config descriptor
	configReader, err := repo.Fetch(ctx, manifest.Config)
	if err != nil {
		return nil, microerror.Maskf(couldNotFetchConfigBlobError, "error fetching config blob: %v", err)
	}
	defer configReader.Close()

	// Read the config content
	configBytes, err := io.ReadAll(configReader)
	if err != nil {
		return nil, microerror.Maskf(couldNotReadConfigError, "error reading config: %v", err)
	}

	// Parse the config as a map
	var configMap map[string]interface{}
	if err := json.Unmarshal(configBytes, &configMap); err != nil {
		return nil, microerror.Maskf(couldNotUnmarshalConfigError, "error unmarshalling config: %v", err)
	}

	return configMap, nil
}
