// Package crdconfig provides functionality to parse CRD configuration files.
package crdconfig

import (
	"fmt"
	"io"
	"os"

	"github.com/giantswarm/microerror"
	"go.yaml.in/yaml/v3"
)

// Item represents a single CRD configuration entry.
type Item struct {
	// URL is the GitHub URL to the CRD YAML file (required).
	URL string `yaml:"url"`

	// Owner is the Backstage owner reference (required).
	Owner string `yaml:"owner"`

	// Lifecycle is the lifecycle stage (optional, defaults to "production").
	Lifecycle string `yaml:"lifecycle"`

	// System is the optional system reference.
	System string `yaml:"system"`
}

// Config holds the service configuration.
type Config struct {
	// Reader is the source to read configuration from.
	// If nil, FilePath must be set.
	Reader io.Reader

	// FilePath is the path to the configuration file.
	// Used if Reader is nil.
	FilePath string
}

// Service provides CRD configuration parsing functionality.
type Service struct {
	config Config
}

// New creates a new CRD configuration service.
func New(c Config) (*Service, error) {
	if c.Reader == nil && c.FilePath == "" {
		return nil, microerror.Maskf(invalidConfigError, "either Reader or FilePath must be provided")
	}

	return &Service{
		config: c,
	}, nil
}

// Load reads and parses the CRD configuration.
func (s *Service) Load() ([]Item, error) {
	var reader io.Reader

	if s.config.Reader != nil {
		reader = s.config.Reader
	} else {
		file, err := os.Open(s.config.FilePath)
		if err != nil {
			return nil, microerror.Maskf(fileNotFoundError, "failed to open config file: %v", err)
		}
		defer file.Close()
		reader = file
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, microerror.Maskf(readError, "failed to read config: %v", err)
	}

	var items []Item
	if err := yaml.Unmarshal(data, &items); err != nil {
		return nil, microerror.Maskf(parseError, "failed to parse YAML: %v", err)
	}

	// Validate and apply defaults
	for i := range items {
		if err := validateItem(&items[i]); err != nil {
			return nil, microerror.Maskf(validationError, "item %d: %v", i+1, err)
		}
		applyDefaults(&items[i])
	}

	return items, nil
}

// validateItem checks that required fields are present.
func validateItem(item *Item) error {
	if item.URL == "" {
		return fmt.Errorf("url is required")
	}
	if item.Owner == "" {
		return fmt.Errorf("owner is required")
	}
	return nil
}

// applyDefaults sets default values for optional fields.
func applyDefaults(item *Item) {
	if item.Lifecycle == "" {
		item.Lifecycle = "production"
	}
}
