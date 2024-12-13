package resource

import (
	"sort"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// Returns an entity representation of the Resource.
func (r *Resource) ToEntity() *bscatalog.Entity {
	sort.Strings(r.Tags)

	e := &bscatalog.Entity{
		APIVersion: bscatalog.APIVersion,
		Kind:       bscatalog.EntityKindResource,
		Metadata: bscatalog.EntityMetadata{
			Annotations: make(map[string]string),
			Description: r.Description,
			Labels:      make(map[string]string),
			Links:       make([]bscatalog.EntityLink, 0),
			Name:        r.Name,
			Title:       r.Title,
		},
	}

	if r.Namespace != "default" {
		e.Metadata.Namespace = r.Namespace
	}

	if len(r.Annotations) > 0 {
		e.Metadata.Annotations = make(map[string]string)
		for k, v := range r.Annotations {
			e.Metadata.Annotations[k] = v
		}
	}

	if len(r.Labels) > 0 {
		e.Metadata.Labels = make(map[string]string)
		for k, v := range r.Labels {
			e.Metadata.Labels[k] = v
		}
	}

	if len(r.Links) > 0 {
		e.Metadata.Links = r.Links
	}

	if r.Tags != nil {
		e.Metadata.Tags = r.Tags
	}

	spec := bscatalog.ResourceSpec{
		Type:  r.Type,
		Owner: r.Owner,
	}
	if r.System != "" {
		spec.System = r.System
	}
	if len(r.DependsOn) > 0 {
		sort.Strings(r.DependsOn)
		spec.DependsOn = r.DependsOn
	}

	e.Metadata.NormalizeTags()
	e.Spec = spec

	return e
}
