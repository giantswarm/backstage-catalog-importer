package catalog

// UserSpec contains the User standard spec fields.
//
// See: https://backstage.io/docs/features/software-catalog/descriptor-format#kind-user
type UserSpec struct {
	// User details.
	Profile UserProfile `yaml:"profile"`

	// Groups the user is a member of. May be empty.
	MemberOf []string `yaml:"memberOf"`
}

type UserProfile struct {
	DisplayName string `yaml:"displayName,omitempty"`
	Email       string `yaml:"email,omitempty"`
	Picture     string `yaml:"picture,omitempty"`
}
