package v1alpha1

type ResourceSpec struct {
	Owner     string   `yaml:"owner"`
	Type      string   `yaml:"type"`
	System    string   `yaml:"system,omitempty"`
	DependsOn []string `yaml:"dependsOn,omitempty"`
}
