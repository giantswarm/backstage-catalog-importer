// Package component provides high level structs and functions to create catalog entities
// of type "component" in an opinionated way. It supports some well-known annotations and labels.
package component

import (
	"fmt"
	"time"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// Component holds our internal representation of something that we want
// to export as a Backstage entity of type "component".
type Component struct {
	// Component name (required)
	Name string

	// Namespace, defaults to "default"
	Namespace string

	// Display title of the component
	Title string

	// Owner of the component. Defaults to "unspecified".
	Owner string

	// Description of the component
	Description string

	// Name of the GitHub repository of the component, in the form of
	// "<organization>/<repository>"
	GithubProjectSlug string

	// Name of the Quay repository of the component, in the form of
	// "<namespace>/<repository>"
	QuayRepositorySlug string

	// Name of the GitHub team owning the component
	// TODO: Specify whether organization name must be prefixed
	GithubTeamSlug string

	// Name of the OpsGenie team owning the component
	OpsGenieTeam string

	// OpsGenie lookup query to find alerts and incidents for one component,
	// similar too 'detailsPair(app:myComponent)'
	OpsGenieComponentSelector string

	// System that the component belongs to
	System string

	// If the project has a CircleCI configuration, name of the project
	CircleCiSlug string

	// If the component type is "service", the 'backstage.io/kubernetes-id' annotation
	// will be set to this value. If empty, the component name will be used.
	KubernetesID string

	// Whether the component repository has a README file
	HasReadme bool

	// Default branch of the component repository
	DefaultBranch string

	// Time of the latest release of the component
	LatestReleaseTime time.Time

	// Tag of the latest release of the component
	LatestReleaseTag string

	// Names to use for Kubernetes resource lookup.
	DeploymentNames []string

	// Component type. Defaults to "unspecified".
	Type string

	// Component lifecycle styge. Defaults to "production".
	Lifecycle string

	// Names of components that this component depends on.
	DependsOn []string

	Tags []string

	Labels map[string]string

	// Extra annotations that are not covered by the fields above.
	Annotations map[string]string

	Links []bscatalog.EntityLink

	Spec bscatalog.ComponentSpec
}

func New(name string, options ...Option) (*Component, error) {
	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	c := &Component{
		Name:      name,
		Namespace: "default",
		Type:      "unspecified",
		Owner:     "unspecified",
		Lifecycle: "production",
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

// Add a tag to the Component
func (c *Component) AddTag(tag string) {
	c.Tags = append(c.Tags, tag)
}

// Add an entity link to the Component
func (c *Component) AddLink(link bscatalog.EntityLink) {
	c.Links = append(c.Links, link)
}

// Set an annotation on the Component. This does not touch the existing annotations,
// but overwrites the value if the key already exists.
func (c *Component) SetAnnotation(key, value string) {
	if c.Annotations == nil {
		c.Annotations = make(map[string]string)
	}
	c.Annotations[key] = value
}

// Set a label on the Component. This does not touch the existing labels,
// but overwrites the value if the key already exists.
func (c *Component) SetLabel(key, value string) {
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	c.Labels[key] = value
}
