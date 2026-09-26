package cmd

import (
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/architectorb"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/repositories"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/component"
)

// orbPins is the one thing applyBuildToolchain needs from the orb resolver,
// kept as a function so the wiring can be tested without GitHub.
type orbPins func(version string) architectorb.Pins

// applyBuildToolchain sets the build toolchain labels and annotations on c
// from what the repo's CircleCI config declares; the vocabulary is in
// pkg/output/catalog/component/toolchain.go.
//
// Nothing is set for a repo that does not use the architect orb. ABS is
// reported only for repos that run push-to-app-catalog, ATS only for repos
// that run run-tests-with-ats, because a version nobody runs is not a fact
// about the repo. Wherever a used tool gets no version label, its status
// label says why.
//
// A repo on a non-release orb ref (dev:<sha>, volatile) gets the raw ref as an
// annotation: there is no orb tag to resolve pins from, and the ref itself is
// not a valid label value. Its ABS is then unknown, and so is its ATS unless
// the repo names the tag itself.
func applyBuildToolchain(c *component.Component, ci repositories.CircleCIConfigDetails, pins orbPins) {
	if ci.ArchitectOrbRef == "" {
		if ci.Incomplete {
			c.SetLabel(component.LabelArchitectOrbStatus, component.ToolchainStatusUnknown)
		}

		return
	}

	resolved := architectorb.Pins{}
	if version, isRelease := architectorb.ReleaseVersion(ci.ArchitectOrbRef); isRelease {
		c.SetLabel(component.LabelArchitectOrbVersion, version)
		resolved = pins(version)
	} else {
		c.SetAnnotation(component.AnnotationArchitectOrbRef, ci.ArchitectOrbRef)
		c.SetLabel(component.LabelArchitectOrbStatus, component.ToolchainStatusNonRelease)
	}

	switch {
	case ci.UsesPushToAppCatalog:
		if !setLabelIfValid(c, component.LabelAppBuildSuiteVersion, resolved.AppBuildSuite) {
			c.SetLabel(component.LabelAppBuildSuiteStatus, component.ToolchainStatusUnknown)
		}
	case ci.Incomplete:
		c.SetLabel(component.LabelAppBuildSuiteStatus, component.ToolchainStatusUnknown)
	}

	switch {
	case ci.UsesRunTestsWithATS && ci.ATSContainerTagConflict:
		c.SetLabel(component.LabelAppTestSuiteStatus, component.ToolchainStatusConflict)
	case ci.UsesRunTestsWithATS:
		tag, source := ci.ATSContainerTag, component.ToolchainATSSourceRepo
		if tag == "" {
			tag, source = resolved.AppTestSuite, component.ToolchainATSSourceOrbDefault
		}
		if setLabelIfValid(c, component.LabelAppTestSuiteVersion, tag) {
			c.SetAnnotation(component.AnnotationAppTestSuiteVersionSource, source)
		} else {
			c.SetLabel(component.LabelAppTestSuiteStatus, component.ToolchainStatusUnknown)
		}
	case ci.Incomplete:
		c.SetLabel(component.LabelAppTestSuiteStatus, component.ToolchainStatusUnknown)
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
