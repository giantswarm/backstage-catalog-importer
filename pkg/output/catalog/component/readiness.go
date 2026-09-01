package component

import (
	"sort"
	"strings"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/helmchart"
	componentutil "github.com/giantswarm/backstage-catalog-importer/pkg/util/component"
)

// Annotation namespaces for chart metadata. The first is the current standard,
// the second the one it replaced. Both are still in use, and app-build-suite's
// HasTeamLabel check accepts either, which is why the migration has stalled: at
// the time of writing 173 charts are on the current prefix, 49 on the legacy one
// and 23 carry both.
const (
	currentAnnotationPrefix = "io.giantswarm.application."
	legacyAnnotationPrefix  = "application.giantswarm.io/"
)

// Chart metadata styles, reported as-is in giantswarm.io/chart-metadata-style.
const (
	MetadataStyleCurrent = "current"
	MetadataStyleLegacy  = "legacy"
	MetadataStyleMixed   = "mixed"
	MetadataStyleNone    = "none"
)

// Enforced gaps. These fail a build today, in app-build-suite inside
// architect/push-to-app-catalog, so a chart carrying one of them cannot be
// released as it stands. They are rare: across 246 charts, one is missing a team
// annotation and three ship no values schema.
const (
	// FlagNoTeam means no team annotation under either prefix. Fails
	// app-build-suite C0001 HasTeamLabel.
	FlagNoTeam = "META-NO-TEAM"
	// FlagNoValuesSchema means a deployable chart ships no values.schema.json.
	// Fails app-build-suite F0001 HasValuesSchema.
	FlagNoValuesSchema = "NO-VALUES-SCHEMA"
)

// Advisory gaps. These are documented in the chart metadata standard and gated
// nowhere: no CI check, no admission policy, nothing. Most of them describe a
// rollout that never happened rather than anything wrong with a given chart —
// 80% of charts have no keywords, 65% no managed annotation, 54% no audience —
// so consumers must render them as adoption, never as failure.
const (
	// FlagMetadataLegacy means a chart still uses the replaced annotation prefix.
	FlagMetadataLegacy = "META-LEGACY"
	// FlagNoAudience means the audience annotation is absent under either prefix.
	FlagNoAudience = "META-NO-AUDIENCE"
	// FlagNoManaged means the managed annotation is absent under either prefix.
	FlagNoManaged = "META-NO-MANAGED"
	// FlagChartAPIV1 means apiVersion is not v2.
	FlagChartAPIV1 = "CHART-API-V1"
	// FlagNoDescription means the required description field is absent.
	FlagNoDescription = "CHART-NO-DESCRIPTION"
	// FlagNoKeywords means the required keywords field is absent.
	FlagNoKeywords = "CHART-NO-KEYWORDS"
	// FlagNoHome means the required home field is absent.
	FlagNoHome = "CHART-NO-HOME"
	// FlagHomeNotGiantSwarm means home is set but does not point at a Giant
	// Swarm repository, which charts vendored from upstream often do.
	FlagHomeNotGiantSwarm = "HOME-NOT-GS"
)

// giantSwarmRepoPrefix is what the standard requires chart home fields to start
// with. Documented only on the internal migration page.
const giantSwarmRepoPrefix = "https://github.com/giantswarm/"

// Standards is the outcome of checking a component's charts against the Giant
// Swarm chart metadata standard.
//
// Enforced and Advisory are kept apart deliberately. Merging them produces a
// verdict that is true of four charts in five, which tells a reader nothing and
// makes a fleet-wide rollout look like hundreds of broken apps.
type Standards struct {
	// Style names the annotation namespace the charts use.
	Style string
	// Enforced lists gaps that fail a build today, de-duplicated and sorted.
	Enforced []string
	// Advisory lists documented-but-ungated gaps, de-duplicated and sorted.
	Advisory []string
}

// Complete reports whether the charts pass everything that is actually
// enforced. It deliberately ignores advisory gaps.
func (s Standards) Complete() bool {
	return len(s.Enforced) == 0
}

// EnforcedString renders the enforced gaps for an annotation value.
func (s Standards) EnforcedString() string {
	return strings.Join(s.Enforced, ",")
}

// AdvisoryString renders the advisory gaps for an annotation value.
func (s Standards) AdvisoryString() string {
	return strings.Join(s.Advisory, ",")
}

// CheckStandards checks a component's charts against the chart metadata
// standard.
//
// chartDirs names the helm/ sub-directory each chart was loaded from, aligned
// with charts by index. It is not the same thing as the chart's own name: a
// chart in helm/foo is free to declare name: bar, and hasValuesSchema is keyed
// by directory, so the two must not be conflated. A chart with no directory
// name has its schema treated as unknown rather than looked up under a key
// that may belong to a different chart.
//
// hasValuesSchema says, per helm/ directory, whether the repo carries
// values.schema.json; a directory absent from the map is treated as unknown
// and never flagged, so an unreadable repo cannot invent a gap.
//
// A gap found in any chart of a multi-chart component is reported for the
// component.
//
// Chart versions are deliberately not looked at here. On a repo's default
// branch they are unsubstituted placeholders (see versions.go), and whether a
// release reached the registry is a question about the registry, not about
// Chart.yaml.
func CheckStandards(charts []*helmchart.Chart, chartDirs []string, hasValuesSchema map[string]bool) Standards {
	if len(charts) == 0 {
		return Standards{Style: MetadataStyleNone}
	}

	enforced := make(map[string]bool)
	advisory := make(map[string]bool)
	sawCurrent := false
	sawLegacy := false

	for i, chart := range charts {
		if chart == nil {
			continue
		}

		current, legacy := annotationStyles(chart.Annotations)
		sawCurrent = sawCurrent || current
		sawLegacy = sawLegacy || legacy

		if legacy && !current {
			advisory[FlagMetadataLegacy] = true
		}

		if !hasAnnotation(chart.Annotations, "team") {
			enforced[FlagNoTeam] = true
		}
		if !hasAnnotation(chart.Annotations, "audience") {
			advisory[FlagNoAudience] = true
		}
		if !hasAnnotation(chart.Annotations, "managed") {
			advisory[FlagNoManaged] = true
		}

		if chart.APIVersion != "v2" {
			advisory[FlagChartAPIV1] = true
		}
		if chart.Description == "" {
			advisory[FlagNoDescription] = true
		}
		if len(chart.Keywords) == 0 {
			advisory[FlagNoKeywords] = true
		}

		// An absent home and a home pointing elsewhere are different problems,
		// and only one of them is somebody's mistake.
		switch {
		case chart.Home == "":
			advisory[FlagNoHome] = true
		case !strings.HasPrefix(chart.Home, giantSwarmRepoPrefix):
			advisory[FlagHomeNotGiantSwarm] = true
		}

		// Library charts have no values to validate, so a missing schema is
		// not a gap for them.
		if componentutil.IsChartDeployable(chart.Type) && i < len(chartDirs) {
			if present, known := hasValuesSchema[chartDirs[i]]; known && !present {
				enforced[FlagNoValuesSchema] = true
			}
		}
	}

	return Standards{
		Style:    metadataStyle(sawCurrent, sawLegacy),
		Enforced: sortedKeys(enforced),
		Advisory: sortedKeys(advisory),
	}
}

// annotationStyles reports which annotation namespaces are present.
func annotationStyles(annotations map[string]string) (current, legacy bool) {
	for key := range annotations {
		if strings.HasPrefix(key, currentAnnotationPrefix) {
			current = true
		}
		if strings.HasPrefix(key, legacyAnnotationPrefix) {
			legacy = true
		}
	}

	return current, legacy
}

// hasAnnotation reports whether the named annotation is set under either
// prefix, with a non-empty value.
func hasAnnotation(annotations map[string]string, name string) bool {
	for _, prefix := range []string{currentAnnotationPrefix, legacyAnnotationPrefix} {
		if value, ok := annotations[prefix+name]; ok && value != "" {
			return true
		}
	}

	return false
}

func metadataStyle(current, legacy bool) string {
	switch {
	case current && legacy:
		return MetadataStyleMixed
	case current:
		return MetadataStyleCurrent
	case legacy:
		return MetadataStyleLegacy
	default:
		return MetadataStyleNone
	}
}

func sortedKeys(set map[string]bool) []string {
	if len(set) == 0 {
		return nil
	}

	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}
