package component

import (
	"strings"

	"github.com/Masterminds/semver/v3"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/ociregistry"
)

// placeholderVersion is what architect leaves in Chart.yaml until it substitutes
// the real version at release time.
const placeholderVersion = "0.0.0"

// isPublishableChartVersion reports whether a chart version is a real release,
// and so worth publishing to the catalog.
//
// The chart catalog (cmd/charts) reads chart versions from the OCI registry and
// applies exactly this rule via ociregistry.IsReleaseVersion. The component
// catalog instead reads helm/<chart>/Chart.yaml from each repo's default branch,
// where architect has not substituted the placeholders yet, so it sees dev
// versions such as 0.0.0-dev or 1.6.0-dev. Those are not versions anybody can
// deploy and must not reach the catalog.
//
// placeholderVersion needs its own check: 0.0.0 is valid semver with no
// prerelease, so IsReleaseVersion accepts it.
func isPublishableChartVersion(version string) bool {
	if version == "" || version == placeholderVersion {
		return false
	}

	return ociregistry.IsReleaseVersion(version)
}

// isPublishableAppVersion reports whether an app version is worth publishing.
//
// This is deliberately narrower than isPublishableChartVersion. An app version
// names the *upstream* release a chart packages and follows no convention: main,
// edge-25.12.3, 4.9, v1.13.2 and 1.12.8-gs-6158079b4 are all real values in the
// catalog today. Requiring pure semver here would discard about fifty legitimate
// app versions to suppress a handful of placeholders, so only values that are
// provably unsubstituted are dropped: the 0.0.0 placeholder, and a semver whose
// prerelease is architect's own dev marker.
func isPublishableAppVersion(version string) bool {
	if version == "" || version == placeholderVersion {
		return false
	}

	v, err := semver.NewVersion(version)
	if err != nil {
		// Not semver at all, so not a placeholder either. Take it at face value.
		return true
	}

	prerelease := v.Prerelease()

	return prerelease != "dev" && !strings.HasPrefix(prerelease, "dev.")
}

// hasAny reports whether at least one of the values carries content. Used to
// decide whether a comma-separated annotation is worth emitting at all: joining
// empty slots yields ",,", which says nothing.
func hasAny(values []string) bool {
	for _, v := range values {
		if v != "" {
			return true
		}
	}

	return false
}
