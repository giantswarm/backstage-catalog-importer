package cmd

import (
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/architectorb"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/repositories"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/component"
)

// Build toolchain, for the devportal's build view. What a repo's default
// branch declares it builds with: the architect orb version, and the
// app-build-suite and app-test-suite versions that orb release pins (or, for
// ATS, the repo's own override). Versions are labels so the catalog can filter
// on them server-side; anything that is not a clean version goes into an
// annotation instead, because Backstage rejects it as a label value.
//
// These say what a build on the default branch would use today. They are not
// proof of what the last build ran: an orb bump that landed after the last
// green build changes the label, not history.
const (
	architectOrbVersionLabel   = "giantswarm.io/architect-orb-version"
	appBuildSuiteVersionLabel  = "giantswarm.io/app-build-suite-version"
	appTestSuiteVersionLabel   = "giantswarm.io/app-test-suite-version"
	architectOrbRefAnnotation  = "giantswarm.io/architect-orb-ref"
	atsVersionSourceAnnotation = "giantswarm.io/app-test-suite-version-source"

	atsVersionSourceRepo       = "repo"
	atsVersionSourceOrbDefault = "orb-default"
)

// orbPins is the one thing applyBuildToolchain needs from the orb resolver,
// kept as a function so the wiring can be tested without GitHub.
type orbPins func(version string) architectorb.Pins

// applyBuildToolchain sets the build toolchain labels and annotations on c
// from what the repo's CircleCI config declares.
//
// Nothing is set for a repo that does not use the architect orb. A repo on a
// non-release orb ref (dev:<sha>, volatile) gets only the raw ref as an
// annotation: there is no orb tag to resolve pins from, and the ref itself is
// not a valid label value. ABS is set only for repos that run
// push-to-app-catalog, ATS only for repos that run run-tests-with-ats, because
// a version nobody runs is not a fact about the repo.
func applyBuildToolchain(c *component.Component, ci repositories.CircleCIConfigDetails, pins orbPins) {
	if ci.ArchitectOrbRef == "" {
		return
	}

	version, isRelease := architectorb.ReleaseVersion(ci.ArchitectOrbRef)
	if !isRelease {
		c.SetAnnotation(architectOrbRefAnnotation, ci.ArchitectOrbRef)

		return
	}

	setLabelIfValid(c, architectOrbVersionLabel, version)

	resolved := pins(version)

	if ci.UsesPushToAppCatalog && resolved.AppBuildSuite != "" {
		setLabelIfValid(c, appBuildSuiteVersionLabel, resolved.AppBuildSuite)
	}

	if ci.UsesRunTestsWithATS && !ci.ATSContainerTagConflict {
		tag, source := ci.ATSContainerTag, atsVersionSourceRepo
		if tag == "" {
			tag, source = resolved.AppTestSuite, atsVersionSourceOrbDefault
		}
		if tag != "" && setLabelIfValid(c, appTestSuiteVersionLabel, tag) {
			c.SetAnnotation(atsVersionSourceAnnotation, source)
		}
	}
}

// setLabelIfValid sets the label only when Backstage would accept the value,
// and reports whether it did. An invalid value is dropped, never mangled: a
// label the catalog cannot filter on is worse than no label.
func setLabelIfValid(c *component.Component, key, value string) bool {
	if !component.IsValidLabelValue(value) {
		return false
	}
	c.SetLabel(key, value)

	return true
}
