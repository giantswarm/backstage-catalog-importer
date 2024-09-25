package group

import (
	"fmt"
	"sort"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// Option is an option to configure a Group.
type Option func(*Group)

// Group holds our internal representation of something that we want
// to export as a Group entity.
type Group struct {
	// Group name (required)
	Name string

	// Namespace, defaults to "default"
	Namespace string

	// Display title of the group
	Title string

	Description              string
	Type                     string
	PictureURL               string
	GrafanaDashboardSelector string
	ChildrenNames            []string
	ParentName               string
	MemberNames              []string
}

func New(name string, options ...Option) (*Group, error) {
	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	c := &Group{
		Name:      name,
		Namespace: "default",
		Type:      "team",
	}

	for _, option := range options {
		option(c)
	}

	return c, nil
}

// Returns an entity representation of the component.
func (c *Group) ToEntity() *bscatalog.Entity {
	sort.Strings(c.MemberNames)
	sort.Strings(c.ChildrenNames)

	spec := bscatalog.GroupSpec{
		Type: c.Type,
		Profile: bscatalog.GroupProfile{
			DisplayName: c.Title,
			Picture:     c.PictureURL,
		},
		Children: c.ChildrenNames,
		Parent:   c.ParentName,
		Members:  c.MemberNames,
	}

	e := &bscatalog.Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       bscatalog.EntityKindGroup,
		Metadata: bscatalog.EntityMetadata{
			Name:        c.Name,
			Description: c.Description,
			Namespace:   c.Namespace,
			Title:       c.Title,
		},
		Spec: spec,
	}

	annotations := map[string]string{}
	if c.GrafanaDashboardSelector != "" {
		annotations["grafana/dashboard-selector"] = c.GrafanaDashboardSelector
	}

	if len(annotations) > 0 {
		e.Metadata.Annotations = annotations
	}

	return e
}
