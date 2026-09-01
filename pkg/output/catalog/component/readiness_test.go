package component

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/helmchart"
)

// compliantChart returns a chart that meets the standard in full, so each test
// can break exactly one thing. Note how rare this shape is in reality: 80% of
// charts in the fleet have no keywords.
func compliantChart(name string) *helmchart.Chart {
	c := &helmchart.Chart{}
	c.Name = name
	c.APIVersion = "v2"
	c.Description = "A compliant chart"
	c.Home = giantSwarmRepoPrefix + name
	c.Keywords = []string{"giantswarm"}
	c.Annotations = map[string]string{
		"io.giantswarm.application.team":     "honeybadger",
		"io.giantswarm.application.audience": "all",
		"io.giantswarm.application.managed":  "true",
	}

	return c
}

func TestCheckStandards(t *testing.T) {
	// chart-operator's real shape at HEAD: current prefix, team only.
	teamOnly := compliantChart("team-only-chart")
	delete(teamOnly.Annotations, "io.giantswarm.application.audience")
	delete(teamOnly.Annotations, "io.giantswarm.application.managed")
	teamOnly.Keywords = nil

	legacy := compliantChart("legacy-chart")
	legacy.Annotations = map[string]string{
		"application.giantswarm.io/team":     "honeybadger",
		"application.giantswarm.io/audience": "all",
		"application.giantswarm.io/managed":  "true",
	}

	mixed := compliantChart("mixed-chart")
	mixed.Annotations["application.giantswarm.io/team"] = "honeybadger"

	bare := compliantChart("bare-chart")
	bare.Annotations = nil

	// A required annotation present but empty counts as absent.
	blankTeam := compliantChart("blank-team-chart")
	blankTeam.Annotations["io.giantswarm.application.team"] = ""

	// Either prefix satisfies the requirement, which is exactly why the legacy
	// prefix survives: app-build-suite's HasTeamLabel accepts both.
	crossPrefix := compliantChart("cross-prefix-chart")
	delete(crossPrefix.Annotations, "io.giantswarm.application.team")
	crossPrefix.Annotations["application.giantswarm.io/team"] = "honeybadger"

	v1 := compliantChart("v1-chart")
	v1.APIVersion = "v1"

	noHome := compliantChart("no-home-chart")
	noHome.Home = ""

	upstreamHome := compliantChart("upstream-home-chart")
	upstreamHome.Home = "https://github.com/kubernetes-sigs/some-project"

	library := compliantChart("library-chart")
	library.Type = "library"

	// A chart whose declared name has nothing to do with the directory it
	// lives in. Real repos do this; schema presence is known per directory.
	renamed := compliantChart("declared-name")

	tests := []struct {
		name   string
		charts []*helmchart.Chart
		// chartDirs is the helm/ directory per chart. Left nil, the runner
		// fills it with each chart's own name, which is the usual case; the
		// cases that exercise a divergence set it explicitly.
		chartDirs       []string
		hasValuesSchema map[string]bool
		want            Standards
	}{
		{
			name:   "no charts",
			charts: nil,
			want:   Standards{Style: MetadataStyleNone},
		},
		{
			name:            "fully compliant",
			charts:          []*helmchart.Chart{compliantChart("good-chart")},
			hasValuesSchema: map[string]bool{"good-chart": true},
			want:            Standards{Style: MetadataStyleCurrent},
		},
		{
			name:            "team only, the common real shape, is enforced-clean",
			charts:          []*helmchart.Chart{teamOnly},
			hasValuesSchema: map[string]bool{"team-only-chart": true},
			want: Standards{
				Style:    MetadataStyleCurrent,
				Advisory: []string{FlagNoKeywords, FlagNoAudience, FlagNoManaged},
			},
		},
		{
			name:            "legacy prefix is advisory, not enforced",
			charts:          []*helmchart.Chart{legacy},
			hasValuesSchema: map[string]bool{"legacy-chart": true},
			want:            Standards{Style: MetadataStyleLegacy, Advisory: []string{FlagMetadataLegacy}},
		},
		{
			name:            "both prefixes is not a legacy gap",
			charts:          []*helmchart.Chart{mixed},
			hasValuesSchema: map[string]bool{"mixed-chart": true},
			want:            Standards{Style: MetadataStyleMixed},
		},
		{
			name:            "no annotations at all fails the enforced team check",
			charts:          []*helmchart.Chart{bare},
			hasValuesSchema: map[string]bool{"bare-chart": true},
			want: Standards{
				Style:    MetadataStyleNone,
				Enforced: []string{FlagNoTeam},
				Advisory: []string{FlagNoAudience, FlagNoManaged},
			},
		},
		{
			name:            "empty team annotation counts as absent",
			charts:          []*helmchart.Chart{blankTeam},
			hasValuesSchema: map[string]bool{"blank-team-chart": true},
			want:            Standards{Style: MetadataStyleCurrent, Enforced: []string{FlagNoTeam}},
		},
		{
			name:            "team under the legacy prefix satisfies the enforced check",
			charts:          []*helmchart.Chart{crossPrefix},
			hasValuesSchema: map[string]bool{"cross-prefix-chart": true},
			want:            Standards{Style: MetadataStyleMixed},
		},
		{
			name:            "apiVersion v1",
			charts:          []*helmchart.Chart{v1},
			hasValuesSchema: map[string]bool{"v1-chart": true},
			want:            Standards{Style: MetadataStyleCurrent, Advisory: []string{FlagChartAPIV1}},
		},
		{
			name:            "home absent and home wrong are different gaps",
			charts:          []*helmchart.Chart{noHome},
			hasValuesSchema: map[string]bool{"no-home-chart": true},
			want:            Standards{Style: MetadataStyleCurrent, Advisory: []string{FlagNoHome}},
		},
		{
			name:            "home points upstream",
			charts:          []*helmchart.Chart{upstreamHome},
			hasValuesSchema: map[string]bool{"upstream-home-chart": true},
			want:            Standards{Style: MetadataStyleCurrent, Advisory: []string{FlagHomeNotGiantSwarm}},
		},
		{
			name:            "no values schema is enforced",
			charts:          []*helmchart.Chart{compliantChart("good-chart")},
			hasValuesSchema: map[string]bool{"good-chart": false},
			want:            Standards{Style: MetadataStyleCurrent, Enforced: []string{FlagNoValuesSchema}},
		},
		{
			name:            "schema presence unknown is never a gap",
			charts:          []*helmchart.Chart{compliantChart("good-chart")},
			hasValuesSchema: nil,
			want:            Standards{Style: MetadataStyleCurrent},
		},
		{
			name:            "library chart needs no values schema",
			charts:          []*helmchart.Chart{library},
			hasValuesSchema: map[string]bool{"library-chart": false},
			want:            Standards{Style: MetadataStyleCurrent},
		},
		{
			name:            "schema is looked up by directory, not declared name",
			charts:          []*helmchart.Chart{renamed},
			chartDirs:       []string{"chart-dir"},
			hasValuesSchema: map[string]bool{"chart-dir": false},
			want:            Standards{Style: MetadataStyleCurrent, Enforced: []string{FlagNoValuesSchema}},
		},
		{
			// The declared name must not reach into a sibling directory's
			// entry, which is what keying by chart.Name did.
			name:            "declared name never picks up another directory's schema",
			charts:          []*helmchart.Chart{renamed},
			chartDirs:       []string{"chart-dir"},
			hasValuesSchema: map[string]bool{"chart-dir": true, "declared-name": false},
			want:            Standards{Style: MetadataStyleCurrent},
		},
		{
			name:            "a chart with no known directory has an unknown schema",
			charts:          []*helmchart.Chart{compliantChart("good-chart")},
			chartDirs:       []string{},
			hasValuesSchema: map[string]bool{"good-chart": false},
			want:            Standards{Style: MetadataStyleCurrent},
		},
		{
			name:            "gaps are unioned across charts and sorted",
			charts:          []*helmchart.Chart{legacy, v1},
			hasValuesSchema: map[string]bool{"legacy-chart": true, "v1-chart": false},
			want: Standards{
				Style:    MetadataStyleMixed,
				Enforced: []string{FlagNoValuesSchema},
				Advisory: []string{FlagChartAPIV1, FlagMetadataLegacy},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chartDirs := tt.chartDirs
			if chartDirs == nil {
				for _, chart := range tt.charts {
					chartDirs = append(chartDirs, chart.Name)
				}
			}

			got := CheckStandards(tt.charts, chartDirs, tt.hasValuesSchema)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CheckStandards() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestStandards_Complete(t *testing.T) {
	// Advisory gaps must never make a component incomplete: 80% of the fleet
	// has one, and none of them fails a build.
	advisoryOnly := Standards{Style: MetadataStyleCurrent, Advisory: []string{FlagNoKeywords}}
	if !advisoryOnly.Complete() {
		t.Error("advisory-only Standards should be complete")
	}
	if (Standards{Enforced: []string{FlagNoTeam}}).Complete() {
		t.Error("Standards with an enforced gap should not be complete")
	}
}

func TestStandards_Strings(t *testing.T) {
	s := Standards{
		Enforced: []string{FlagNoTeam, FlagNoValuesSchema},
		Advisory: []string{FlagChartAPIV1, FlagMetadataLegacy},
	}
	if got, want := s.EnforcedString(), "META-NO-TEAM,NO-VALUES-SCHEMA"; got != want {
		t.Errorf("EnforcedString() = %q, want %q", got, want)
	}
	if got, want := s.AdvisoryString(), "CHART-API-V1,META-LEGACY"; got != want {
		t.Errorf("AdvisoryString() = %q, want %q", got, want)
	}
	if got := (Standards{}).AdvisoryString(); got != "" {
		t.Errorf("AdvisoryString() on empty = %q, want empty", got)
	}
}
