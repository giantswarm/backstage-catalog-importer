package v1alpha1

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
