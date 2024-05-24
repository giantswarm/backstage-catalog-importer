package catalog

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Component holds our internal representation of something that we want
// to export as a Component entity.
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

	Links []EntityLink

	Spec ComponentSpec
}

// ComponentSpec contains the spec fields for a component entity in Backstage.
//
// See: https://backstage.io/docs/features/software-catalog/descriptor-format/#kind-component
type ComponentSpec struct {
	// The type of component. This field is required.
	Type string `yaml:"type"`

	// The lifecycle state of the component. This field is required.
	Lifecycle string `yaml:"lifecycle"`

	// An entity reference to the owner of the component. This field is required.
	Owner string `yaml:"owner"`

	// An entity reference to the system that the component belongs to.
	System string `yaml:"system,omitempty"`

	// An entity reference to another component of which the component is a part.
	SubcomponentOf string `yaml:"subcomponentOf,omitempty"`

	// An array of entity references to the APIs that are provided by the component.
	ProvidesAPIs []string `yaml:"providesApis,omitempty"`

	// An array of entity references to the APIs that are consumed by the component.
	ConsumesAPIs []string `yaml:"consumesApis,omitempty"`

	// An array of entity references to the components and resources that the component depends on.
	DependsOn []string `yaml:"dependsOn,omitempty"`
}

func (c *Component) AddTag(tag string) {
	c.Tags = append(c.Tags, tag)
}

func (c *Component) AddLink(link EntityLink) {
	c.Links = append(c.Links, link)
}

func (c *Component) SetAnnotation(key, value string) {
	if c.Annotations == nil {
		c.Annotations = make(map[string]string)
	}
	c.Annotations[key] = value
}

func (c *Component) SetLabel(key, value string) {
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	c.Labels[key] = value
}

// Returns an entity representation of the component.
func (c *Component) ToEntity() *Entity {
	sort.Strings(c.Tags)

	e := &Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       EntityKindComponent,
		Metadata: EntityMetadata{
			Annotations: make(map[string]string),
			Description: c.Description,
			Labels:      make(map[string]string),
			Links:       make([]EntityLink, 0),
			Name:        c.Name,
			Title:       c.Title,
		},
	}

	if c.Namespace != "default" {
		e.Metadata.Namespace = c.Namespace
	}

	if c.Annotations != nil {
		e.Metadata.Annotations = make(map[string]string)
		for k, v := range c.Annotations {
			e.Metadata.Annotations[k] = v
		}
	}

	if c.Labels != nil {
		e.Metadata.Labels = make(map[string]string)
		for k, v := range c.Labels {
			e.Metadata.Labels[k] = v
		}
	}

	if c.Links != nil {
		e.Metadata.Links = c.Links
	}

	if c.Tags != nil {
		e.Metadata.Tags = c.Tags
	}

	if c.GithubProjectSlug != "" {
		e.Metadata.Annotations["github.com/project-slug"] = c.GithubProjectSlug
		e.Metadata.Annotations["backstage.io/source-location"] = fmt.Sprintf("url:https://github.com/%s", c.GithubProjectSlug)
		if c.HasReadme && c.DefaultBranch != "" {
			e.Metadata.Annotations["backstage.io/techdocs-ref"] = fmt.Sprintf("url:https://github.com/%s/tree/%s", c.GithubProjectSlug, c.DefaultBranch)
		}
	}
	if c.QuayRepositorySlug != "" {
		e.Metadata.Annotations["quay.io/repository-slug"] = c.QuayRepositorySlug
	}
	if c.GithubTeamSlug != "" {
		e.Metadata.Annotations["github.com/team-slug"] = c.GithubTeamSlug
	}
	if c.OpsGenieTeam != "" {
		e.Metadata.Annotations["opsgenie.com/team"] = c.OpsGenieTeam
	}
	if c.OpsGenieComponentSelector != "" {
		e.Metadata.Annotations["opsgenie.com/component-selector"] = c.OpsGenieComponentSelector
	}
	if c.LatestReleaseTag != "" {
		e.Metadata.Annotations["giantswarm.io/latest-release-tag"] = c.LatestReleaseTag
	}
	if c.LatestReleaseTime.Format(time.RFC3339) != "0001-01-01T00:00:00Z" {
		e.Metadata.Annotations["giantswarm.io/latest-release-date"] = c.LatestReleaseTime.Format(time.RFC3339)
	}
	if c.CircleCiSlug != "" {
		e.Metadata.Annotations["circleci.com/project-slug"] = c.CircleCiSlug
	}
	if c.DeploymentNames != nil {
		sort.Strings(c.DeploymentNames)
		e.Metadata.Annotations["giantswarm.io/deployment-names"] = strings.Join(c.DeploymentNames, ",")
	}
	if c.Type == "service" {
		e.Metadata.Annotations["backstage.io/kubernetes-id"] = c.Name
		if c.KubernetesID != "" {
			e.Metadata.Annotations["backstage.io/kubernetes-id"] = c.KubernetesID
		}
	}

	spec := ComponentSpec{
		Type:      c.Type,
		Lifecycle: c.Lifecycle,
		Owner:     c.Owner,
	}
	if c.System != "" {
		spec.System = c.System
	}
	if len(c.DependsOn) > 0 {
		sort.Strings(c.DependsOn)
		for i, d := range c.DependsOn {
			c.DependsOn[i] = "component:" + d
		}
		spec.DependsOn = c.DependsOn
	}

	e.Spec = spec

	return e
}
