package repositories

type Repo struct {
	Name            string           `yaml:"name"`
	ComponentType   string           `yaml:"componentType"`
	DeploymentNames []string         `yaml:"deploymentNames"`
	System          string           `yaml:"system"`
	Gen             RepoGen          `yaml:"gen"`
	Lifecycle       RepoLifecycle    `yaml:"lifecycle"`
	Replacements    RepoReplacements `yaml:"replace"`
	AppTestSuite    interface{}      `yaml:"app_test_suite"`
}

type RepoFlavor string
type RepoLanguage string
type RepoLifecycle string

const (
	RepoFlavorApp     RepoFlavor = "app"
	RepoFlavorCLI     RepoFlavor = "cli"
	RepoFlavorGeneric RepoFlavor = "generic"
)

const (
	RepoLanguageGo      RepoLanguage = "go"
	RepoLanguagePython  RepoLanguage = "python"
	RepoLanguageGeneric RepoLanguage = "generic"
)

type RepoGen struct {
	Flavors                 []RepoFlavor `yaml:"flavours"`
	Language                RepoLanguage `yaml:"language"`
	InstallUpdateChart      bool         `yaml:"installUpdateChart"`
	EnableFloatingMajorTags bool         `yaml:"enableFloatingMajorTags"`
}

type RepoReplacements struct {
	ArchitectOrb     bool `yaml:"architect-orb"`
	Renovate         bool `yaml:"renovate"`
	PreCommit        bool `yaml:"precommit"`
	DependabotRemove bool `yaml:"dependabotRemove"`
}
