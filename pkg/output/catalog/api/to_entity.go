package api

import (
	"maps"

	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

// ToEntity converts the API to a Backstage entity.
func (a *API) ToEntity() *bscatalog.Entity {
	e := &bscatalog.Entity{
		APIVersion: bscatalog.APIVersion,
		Kind:       bscatalog.EntityKindAPI,
		Metadata: bscatalog.EntityMetadata{
			Name:        a.Name,
			Title:       a.Title,
			Description: a.Description,
			Tags:        a.Tags,
			Links:       a.Links,
		},
	}

	// Only set namespace if not "default"
	if a.Namespace != "default" {
		e.Metadata.Namespace = a.Namespace
	}

	// Copy annotations
	if len(a.Annotations) > 0 {
		e.Metadata.Annotations = make(map[string]string)
		maps.Copy(e.Metadata.Annotations, a.Annotations)
	}

	// Copy labels
	if len(a.Labels) > 0 {
		e.Metadata.Labels = make(map[string]string)
		maps.Copy(e.Metadata.Labels, a.Labels)
	}

	// Build the spec
	spec := bscatalog.APISpec{
		Type:       a.Type,
		Lifecycle:  a.Lifecycle,
		Owner:      a.Owner,
		Definition: a.Definition,
	}

	if a.System != "" {
		spec.System = a.System
	}

	e.Spec = spec

	// Normalize tags (lowercase, replace special chars)
	e.Metadata.NormalizeTags()

	return e
}
