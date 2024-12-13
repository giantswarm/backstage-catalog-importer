package v1alpha1

import "strings"

// An Entity in the software catalog.
type Entity struct {
	// APIVersion is the version of specification format for this particular entity.
	APIVersion string `yaml:"apiVersion"`

	// Kind is the high-level entity type.
	Kind EntityKind `yaml:"kind"`

	// Metadata related to the entity.
	Metadata EntityMetadata `yaml:"metadata"`

	// Relations that this entity has with other entities.
	Relations []EntityRelation `yaml:"relations,omitempty"`

	Spec interface{}
}

// EntityMetadata contains fields common to all versions/kinds of entity.
//
// See also:
//
// https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta
// https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/
type EntityMetadata struct {
	// The Name of the entity.
	//
	// Must be unique within the catalog at any given point in time, for any
	// given namespace + kind pair. This value is part of the technical
	// identifier of the entity, and as such it will appear in URLs, database
	// tables, entity references, and similar. It is subject to restrictions
	// regarding what characters are allowed.
	//
	// If you want to use a different, more human-readable string with fewer
	// restrictions on it in user interfaces, see the `title` field below.
	Name string `yaml:"name"`

	// The Namespace that the entity belongs to.
	Namespace string `yaml:"namespace,omitempty"`

	// Title is a display name of the entity, to be presented in user interfaces instead
	// of the `name` property above, when available.
	//
	// This field is sometimes useful when the `name` is cumbersome or ends up
	// being perceived as overly technical. The title generally does not have
	// as stringent format requirements on it, so it may contain special
	// characters and be more explanatory. Do keep it very short though, and
	// avoid situations where a title can be confused with the name of another
	// entity, or where two entities share a title.
	//
	// Note that this is only for display purposes, and may be ignored by some
	// parts of the code. Entity references still always make use of the `name`
	// property, not the title.
	Title string `yaml:"title,omitempty"`

	// Description is a short (typically relatively few words, on one line) description of the entity.
	Description string `yaml:"description,omitempty"`

	// Labels contains key/value pairs of identifying information attached to the entity.
	Labels map[string]string `yaml:"labels,omitempty"`

	// Annotations contains key/value pairs of non-identifying auxiliary information attached to the entity.
	Annotations map[string]string `yaml:"annotations,omitempty"`

	// Tags is a list of single-valued strings, to for example classify catalog entities in various ways.
	Tags []string `yaml:"tags,omitempty"`

	// Links is a list of external hyperlinks related to the entity.
	Links []EntityLink `yaml:"links,omitempty"`
}

// EntityRelation is a relation of a specific type to another entity in the catalog.
type EntityRelation struct {
	// The type of the relation.
	Type string `yaml:"type"`

	// The entity ref of the target of this relation.
	TargetRef string `yaml:"targetRef"`
}

// EntityLink is a link to external information that is related to the entity.
type EntityLink struct {
	// URL to the external site, document, etc.
	URL string `yaml:"url"`

	// Title is an optional descriptive title for the link.
	Title string `yaml:"title,omitempty"`

	// Icon is an optional semantic key that represents a visual icon.
	Icon string `yaml:"icon,omitempty"`

	// Type is an optional value to categorize links into specific groups.
	Type string `yaml:"type,omitempty"`
}

// NormalizeTags normalizes the tags of the entity. Tags may only use the characters
// a-z, 0-9, and -
func (e *EntityMetadata) NormalizeTags() {
	if e.Tags == nil {
		return
	}

	for i := range e.Tags {
		e.Tags[i] = strings.ToLower(e.Tags[i])
		e.Tags[i] = strings.TrimSpace(e.Tags[i])
		e.Tags[i] = strings.ReplaceAll(e.Tags[i], " ", "-")
		e.Tags[i] = strings.ReplaceAll(e.Tags[i], "_", "-")
		e.Tags[i] = strings.ReplaceAll(e.Tags[i], "/", "-")
		e.Tags[i] = strings.ReplaceAll(e.Tags[i], "+", "-")
	}
}
