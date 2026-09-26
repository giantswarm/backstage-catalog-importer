package component

// Build toolchain, for the devportal's build view: what a repo's default
// branch declares it builds with. Versions are labels so the catalog can
// filter on them server-side; anything that is not a clean version goes into
// an annotation instead, because Backstage rejects it as a label value.
//
// These say what a build on the default branch would use today. They are not
// proof of what the last build ran: an orb bump that landed after the last
// green build changes the label, not history.
const (
	LabelArchitectOrbVersion  = "giantswarm.io/architect-orb-version"
	LabelAppBuildSuiteVersion = "giantswarm.io/app-build-suite-version"
	LabelAppTestSuiteVersion  = "giantswarm.io/app-test-suite-version"

	// AnnotationArchitectOrbRef carries a non-release orb ref (dev:<sha>,
	// volatile), which is not a valid label value.
	AnnotationArchitectOrbRef = "giantswarm.io/architect-orb-ref"

	// AnnotationAppTestSuiteVersionSource says where the ATS version came
	// from: ToolchainATSSourceRepo or ToolchainATSSourceOrbDefault.
	AnnotationAppTestSuiteVersionSource = "giantswarm.io/app-test-suite-version-source"

	ToolchainATSSourceRepo       = "repo"
	ToolchainATSSourceOrbDefault = "orb-default"
)

// Status labels say why a tool the repo uses has no version label, so that
// "could not tell" is a filter of its own instead of a repo silently missing
// from every version filter. Each is set only when the matching version label
// is not; a repo that does not use a tool gets neither.
const (
	LabelArchitectOrbStatus  = "giantswarm.io/architect-orb-status"
	LabelAppBuildSuiteStatus = "giantswarm.io/app-build-suite-status"
	LabelAppTestSuiteStatus  = "giantswarm.io/app-test-suite-status"

	// ToolchainStatusNonRelease: the orb is referenced by a non-release ref,
	// see AnnotationArchitectOrbRef. Architect orb only.
	ToolchainStatusNonRelease = "non-release"
	// ToolchainStatusConflict: the repo would run more than one version.
	// app-test-suite only.
	ToolchainStatusConflict = "conflict"
	// ToolchainStatusUnknown: the tool is (or may be) used, but its version
	// could not be determined — the config or the orb source could not be
	// read, the orb ref is not a release, or the value is not a version.
	ToolchainStatusUnknown = "unknown"
)
