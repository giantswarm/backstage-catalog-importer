package group

import (
	"fmt"
	"sort"

	"github.com/giantswarm/backstage-catalog-importer/pkg/catalog"
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
	OpsgenieTeamName         string
	ChildrenNames            []string
	ParentName               string
	MemberNames              []string
}

func NewGroup(name string, options ...Option) (*Group, error) {
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
func (c *Group) ToEntity() *catalog.Entity {
	sort.Strings(c.MemberNames)
	sort.Strings(c.ChildrenNames)

	spec := catalog.GroupSpec{
		Type: c.Type,
		Profile: catalog.GroupProfile{
			DisplayName: c.Title,
			Picture:     c.PictureURL,
		},
		Children: c.ChildrenNames,
		Parent:   c.ParentName,
		Members:  c.MemberNames,
	}

	e := &catalog.Entity{
		APIVersion: "backstage.io/v1alpha1",
		Kind:       catalog.EntityKindGroup,
		Metadata: catalog.EntityMetadata{
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
	if c.OpsgenieTeamName != "" {
		annotations["opsgenie.io/team-name"] = c.OpsgenieTeamName
	}

	if len(annotations) > 0 {
		e.Metadata.Annotations = annotations
	}

	return e
}
