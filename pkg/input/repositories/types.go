package repositories

// Repo represents an entry in the giantswarm/github repositories YAML data.
type Repo struct {
	Name          string             `yaml:"name"`
	ComponentType string             `yaml:"componentType"`
	System        string             `yaml:"system"`
	Gen           RepoGen            `yaml:"gen"`
	Lifecycle     RepoLifecycle      `yaml:"lifecycle"`
	Replacements  RepoReplacements   `yaml:"replace"`
	AppTestSuite  interface{}        `yaml:"app_test_suite"`
	UpstreamCheck *RepoUpstreamCheck `yaml:"upstreamCheck"`
}

// HasUpstreamCheck reports whether the repo opted into the monthly upstream
// update check.
func (r Repo) HasUpstreamCheck() bool {
	return r.UpstreamCheck != nil
}

// EffectiveReleaseWorkflow returns the release workflow that applies to the
// repo, mirroring the default logic in github's repositories.schema.json: an
// explicit gen.ci.releaseWorkflow wins; otherwise a devctl-generated CI surface
// (gen.ci.generate) implies "auto-release", and everything else is "legacy".
func (r Repo) EffectiveReleaseWorkflow() string {
	if r.Gen.CI.ReleaseWorkflow != "" {
		return string(r.Gen.CI.ReleaseWorkflow)
	}
	if r.Gen.CI.Generate {
		return "auto-release"
	}
	return "legacy"
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
	PreCommit          []string     `yaml:"preCommit"`
	CI                 RepoCI       `yaml:"ci"`
}

// RepoCI holds the gen.ci block: whether devctl generates the CI surface and
// which release workflow it produces.
type RepoCI struct {
	Generate        bool   `yaml:"generate"`
	ReleaseWorkflow string `yaml:"releaseWorkflow"`
}

// RepoUpstreamCheck holds the upstreamCheck block that opts a repo into the
// monthly upstream update check.
type RepoUpstreamCheck struct {
	ChartPath     string `yaml:"chartPath"`
	UpstreamRepo  string `yaml:"upstreamRepo"`
	ReleasePrefix string `yaml:"releasePrefix"`
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

	// Whether the CircleCI config uses force-public in push-to-registries,
	// meaning charts/images are published to the public registry even though
	// the repository is private.
	ForcePublicRegistry bool

	// Whether the repository has a README.md in the root directory.
	HasReadme bool

	// Whether the repository has a "helm" folder in the root directory.
	HasHelmFolder bool

	NumHelmCharts int

	HelmChartNames []string
}
