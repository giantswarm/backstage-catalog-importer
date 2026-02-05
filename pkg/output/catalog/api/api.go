// Package api provides high level structs and functions to create catalog entities
// of type "API" in an opinionated way.
package api

import (
	"fmt"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// API holds our internal representation of something that we want
// to export as a Backstage entity of kind "API".
type API struct {
	// Name is the API name (required). For CRDs, this is the full CRD name
	// (e.g., "apps.application.giantswarm.io").
	Name string

	// Namespace defaults to "default".
	Namespace string

	// Title is the display title of the API.
	Title string

	// Description of the API.
	Description string

	// Owner is the entity reference to the owner (required).
	Owner string

	// Type is the API type (e.g., "crd", "openapi", "grpc").
	Type string

	// Lifecycle is the lifecycle stage (e.g., "production", "experimental").
	Lifecycle string

	// System is an optional reference to the system the API belongs to.
	System string

	// Definition contains the API definition content.
	// For CRDs, this is the full CRD YAML.
	Definition string

	// Tags for categorization.
	Tags []string

	// Labels for key/value metadata.
	Labels map[string]string

	// Annotations for non-identifying metadata.
	Annotations map[string]string

	// Links to external resources.
	Links []bscatalog.EntityLink
}

// New creates a new API with the given name and options.
func New(name string, options ...Option) (*API, error) {
	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	a := &API{
		Name:      name,
		Namespace: "default",
		Type:      "crd",
		Lifecycle: "production",
		Owner:     "unspecified",
	}

	for _, option := range options {
		option(a)
	}

	return a, nil
}

// AddTag adds a tag to the API.
func (a *API) AddTag(tag string) {
	a.Tags = append(a.Tags, tag)
}

// AddLink adds an entity link to the API.
func (a *API) AddLink(link bscatalog.EntityLink) {
	a.Links = append(a.Links, link)
}

// SetAnnotation sets an annotation on the API.
// This overwrites the value if the key already exists.
func (a *API) SetAnnotation(key, value string) {
	if a.Annotations == nil {
		a.Annotations = make(map[string]string)
	}
	a.Annotations[key] = value
}

// SetLabel sets a label on the API.
// This overwrites the value if the key already exists.
func (a *API) SetLabel(key, value string) {
	if a.Labels == nil {
		a.Labels = make(map[string]string)
	}
	a.Labels[key] = value
}
