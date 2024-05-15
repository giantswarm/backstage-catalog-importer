package appcatalog

type Index struct {
	APIVersion string             `yaml:"apiVersion"`
	Entries    map[string][]Entry `yaml:"entries"`
	Generated  string             `yaml:"generated"`
}

type Entry struct {
	Annotations map[string]string `yaml:"annotations"`
	APIVersion  string            `yaml:"apiVersion"`
	AppVersion  string            `yaml:"appVersion"`
	Created     string            `yaml:"created"`
	Description string            `yaml:"description"`
	Digest      string            `yaml:"digest"`
	Home        string            `yaml:"home"`
	Icon        string            `yaml:"icon"`
	Keywords    []string          `yaml:"keywords"`
	Name        string            `yaml:"name"`
	Type        string            `yaml:"type"`
	Urls        []string          `yaml:"urls"`
	Version     string            `yaml:"version"`
}
