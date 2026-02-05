package v1alpha1

// APISpec contains the spec fields for an API entity in Backstage.
// See: https://backstage.io/docs/features/software-catalog/descriptor-format#kind-api
type APISpec struct {
	// Type is the type of API definition (e.g., "openapi", "grpc", "crd").
	Type string `yaml:"type"`

	// Lifecycle is the lifecycle state of the API (e.g., "production", "experimental").
	Lifecycle string `yaml:"lifecycle"`

	// Owner is an entity reference to the owner of the API.
	Owner string `yaml:"owner"`

	// System is an optional entity reference to the system that the API belongs to.
	System string `yaml:"system,omitempty"`

	// Definition contains the actual API definition.
	// For CRDs, this is the full CRD YAML content.
	Definition string `yaml:"definition"`
}
