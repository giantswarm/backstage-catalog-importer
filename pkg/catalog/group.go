package catalog

// GroupSpec contains the Group standard spec fields.
//
// See: https://backstage.io/docs/features/software-catalog/descriptor-format#kind-component
type GroupSpec struct {
	// The type of group. This field is required.
	Type string `yaml:"type"`

	Profile GroupProfile `yaml:"profile,omitempty"`

	// Array of child groups. May be empty.
	Children []string `yaml:"children"`

	// An optional parent group name.
	Parent string `yaml:"parent,omitempty"`

	// Member user names
	Members []string `yaml:"members,omitempty"`
}

type GroupProfile struct {
	DisplayName string `yaml:"displayName,omitempty"`
	Email       string `yaml:"email,omitempty"`
	Picture     string `yaml:"picture,omitempty"`
}
