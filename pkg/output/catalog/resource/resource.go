package resource

import (
	"fmt"
	"time"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

type Resource struct {
	// Resource name (required)
	Name string

	// Namespace, defaults to "default"
	Namespace string

	// Display title of the resource
	Title string

	// Owner of the resource. Defaults to "unspecified".
	Owner string

	// Resource type (required)
	Type string

	// System the resource belongs to
	System string

	// Description of the component
	Description string

	// Creation date/time
	CreatedAt time.Time

	// Names of components that this component depends on.
	DependsOn []string

	Tags []string

	Labels map[string]string

	// Extra annotations that are not covered by the fields above.
	Annotations map[string]string

	Links []bscatalog.EntityLink

	Spec bscatalog.ComponentSpec
}

func New(name string, options ...Option) (*Resource, error) {
	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	c := &Resource{
		Name:      name,
		Namespace: "default",
		Type:      "unspecified",
		Owner:     "unspecified",
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

// Add a tag to the Component
func (c *Resource) AddTag(tag string) {
	c.Tags = append(c.Tags, tag)
}

// Add an entity link to the Component
func (c *Resource) AddLink(link bscatalog.EntityLink) {
	c.Links = append(c.Links, link)
}

// Set an annotation on the Component. This does not touch the existing annotations,
// but overwrites the value if the key already exists.
func (c *Resource) SetAnnotation(key, value string) {
	if c.Annotations == nil {
		c.Annotations = make(map[string]string)
	}
	c.Annotations[key] = value
}

// Set a label on the Component. This does not touch the existing labels,
// but overwrites the value if the key already exists.
func (c *Resource) SetLabel(key, value string) {
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	c.Labels[key] = value
}
