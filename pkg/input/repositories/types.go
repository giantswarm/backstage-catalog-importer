package repositories

import "time"

// Repo represents an entry in the giantswarm/github repositories YAML data.
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
	Flavors            []RepoFlavor `yaml:"flavours"`
	Language           RepoLanguage `yaml:"language"`
	InstallUpdateChart bool         `yaml:"installUpdateChart"`
}

type RepoReplacements struct {
	ArchitectOrb bool `yaml:"architect-orb"`
	Renovate     bool `yaml:"renovate"`
	PreCommit    bool `yaml:"precommit"`
}

// A sparse struct for caching just the GitHub
// repository details we need.
type GithubRepoDetails struct {
	// Repository name
	Name string

	// Repository description
	Description string

	// Name of the default branch
	DefaultBranch string

	// Whether the repository is private. If false, it's public.
	IsPrivate bool

	// The main programming language in the repo.
	MainLanguage string
}

// A struct for caching repository content information.
type GithubRepoContentDetails struct {
	// Whether the repository has a CircleCI configuration file.
	HasCircleCI bool

	// Whether the repository has a README.md in the root directory.
	HasReadme bool

	// Whether the repository has a "helm" folder in the root directory.
	HasHelmFolder bool

	NumHelmCharts int

	HelmChartNames []string
}

// Cache for info on releases of a repo.
type GithubReleaseDetails struct {
	// Creation date/time of the latest release.
	LatestReleaseTime time.Time

	// Tag name of the latest release.
	LatestReleaseTag string
}
