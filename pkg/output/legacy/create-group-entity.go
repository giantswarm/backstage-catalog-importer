package legacy

import (
	"fmt"
	"sort"
	"strings"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// CreateGroupEntity is the deprecated way of generating a group entity.
//
// Deprecated: Use the catalog.Group struct and its ToEntity() method instead.
func CreateGroupEntity(name, displayName, description, parent string, members []string, id int64) bscatalog.Entity {
	sort.Strings(members)
	e := bscatalog.Entity{
		APIVersion: bscatalog.APIVersion,
		Kind:       bscatalog.EntityKindGroup,
		Metadata: bscatalog.EntityMetadata{
			Name: name,
			Annotations: map[string]string{
				"grafana/dashboard-selector": fmt.Sprintf("tags @> 'owner:%s'", name),

				// Like group name, but without "team-" prefix
				"opsgenie.com/team": strings.TrimPrefix(name, "team-"),
			},
		},
	}
	spec := bscatalog.GroupSpec{
		Type:    "team",
		Members: members,
		Profile: bscatalog.GroupProfile{
			Picture: fmt.Sprintf("https://avatars.githubusercontent.com/t/%d?s=116&v=4", id),
		},
	}

	if description != "" {
		e.Metadata.Description = description
	}
	if displayName != "" {
		spec.Profile.DisplayName = displayName
	}
	if parent != "" {
		spec.Parent = parent
	}

	e.Spec = spec

	return e
}
