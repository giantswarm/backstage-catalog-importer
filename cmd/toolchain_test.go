package cmd

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/architectorb"
	"github.com/giantswarm/backstage-catalog-importer/pkg/input/repositories"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/component"
)

func TestApplyBuildToolchain(t *testing.T) {
	pins := func(version string) architectorb.Pins {
		switch version {
		case "10.3.0":
			return architectorb.Pins{AppBuildSuite: "2.3.0", AppTestSuite: "0.15.0"}
		case "4.0.0":
			// Predates run-tests-with-ats.
			return architectorb.Pins{AppBuildSuite: "0.2.4"}
		default:
			return architectorb.Pins{}
		}
	}

	tests := []struct {
		name            string
		ci              repositories.CircleCIConfigDetails
		wantLabels      map[string]string
		wantAnnotations map[string]string
	}{
		{
			name: "no CircleCI or no architect orb sets nothing",
			ci:   repositories.CircleCIConfigDetails{},
		},
		{
			name: "orb only, no chart jobs: just the orb version",
			ci:   repositories.CircleCIConfigDetails{ArchitectOrbRef: "10.3.0"},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-version": "10.3.0",
			},
		},
		{
			name: "chart push and ATS on the orb default",
			ci: repositories.CircleCIConfigDetails{
				ArchitectOrbRef:      "v10.3.0",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-version":   "10.3.0",
				"giantswarm.io/app-build-suite-version": "2.3.0",
				"giantswarm.io/app-test-suite-version":  "0.15.0",
			},
			wantAnnotations: map[string]string{
				"giantswarm.io/app-test-suite-version-source": "orb-default",
			},
		},
		{
			name: "repo overrides the ATS container tag",
			ci: repositories.CircleCIConfigDetails{
				ArchitectOrbRef:      "10.3.0",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
				ATSContainerTag:      "0.4.1",
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-version":   "10.3.0",
				"giantswarm.io/app-build-suite-version": "2.3.0",
				"giantswarm.io/app-test-suite-version":  "0.4.1",
			},
			wantAnnotations: map[string]string{
				"giantswarm.io/app-test-suite-version-source": "repo",
			},
		},
		{
			name: "conflicting ATS overrides emit no ATS version",
			ci: repositories.CircleCIConfigDetails{
				ArchitectOrbRef:         "10.3.0",
				UsesRunTestsWithATS:     true,
				ATSContainerTagConflict: true,
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-version": "10.3.0",
				"giantswarm.io/app-test-suite-status": "conflict",
			},
		},
		{
			name: "ATS job used but the orb predates a default: unknown, not guessed",
			ci: repositories.CircleCIConfigDetails{
				ArchitectOrbRef:      "4.0.0",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-version":   "4.0.0",
				"giantswarm.io/app-build-suite-version": "0.2.4",
				"giantswarm.io/app-test-suite-status":   "unknown",
			},
		},
		{
			name: "orb release whose pins could not be read: versions unknown",
			ci: repositories.CircleCIConfigDetails{
				ArchitectOrbRef:      "99.0.0",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-version":  "99.0.0",
				"giantswarm.io/app-build-suite-status": "unknown",
				"giantswarm.io/app-test-suite-status":  "unknown",
			},
		},
		{
			name: "dev orb ref: raw ref annotated, ABS unknown, repo's own ATS tag still stands",
			ci: repositories.CircleCIConfigDetails{
				ArchitectOrbRef:      "dev:abc123",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
				ATSContainerTag:      "0.4.1",
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-status":   "non-release",
				"giantswarm.io/app-build-suite-status": "unknown",
				"giantswarm.io/app-test-suite-version": "0.4.1",
			},
			wantAnnotations: map[string]string{
				"giantswarm.io/architect-orb-ref":             "dev:abc123",
				"giantswarm.io/app-test-suite-version-source": "repo",
			},
		},
		{
			name: "dev orb ref with ATS on the orb default: ATS unknown",
			ci: repositories.CircleCIConfigDetails{
				ArchitectOrbRef:     "dev:abc123",
				UsesRunTestsWithATS: true,
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-status":  "non-release",
				"giantswarm.io/app-test-suite-status": "unknown",
			},
			wantAnnotations: map[string]string{
				"giantswarm.io/architect-orb-ref": "dev:abc123",
			},
		},
		{
			name: "volatile orb ref",
			ci:   repositories.CircleCIConfigDetails{ArchitectOrbRef: "volatile"},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-status": "non-release",
			},
			wantAnnotations: map[string]string{
				"giantswarm.io/architect-orb-ref": "volatile",
			},
		},
		{
			name: "an ATS override that is not a valid label value is unknown",
			ci: repositories.CircleCIConfigDetails{
				ArchitectOrbRef:     "10.3.0",
				UsesRunTestsWithATS: true,
				ATSContainerTag:     "sha256:abc",
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-version": "10.3.0",
				"giantswarm.io/app-test-suite-status": "unknown",
			},
		},
		{
			name: "an ATS tag given as a CircleCI template expression is unknown",
			ci: repositories.CircleCIConfigDetails{
				ArchitectOrbRef:     "10.3.0",
				UsesRunTestsWithATS: true,
				ATSContainerTag:     "<< parameters.ats_version >>",
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-version": "10.3.0",
				"giantswarm.io/app-test-suite-status": "unknown",
			},
		},
		{
			name: "incomplete config without an orb ref: the orb is unknown",
			ci:   repositories.CircleCIConfigDetails{DynamicSetup: true, Incomplete: true},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-status": "unknown",
			},
		},
		{
			name: "incomplete config with an orb ref: undetected tools are unknown, not absent",
			ci: repositories.CircleCIConfigDetails{
				DynamicSetup:         true,
				Incomplete:           true,
				ArchitectOrbRef:      "10.3.0",
				UsesPushToAppCatalog: true,
			},
			wantLabels: map[string]string{
				"giantswarm.io/architect-orb-version":   "10.3.0",
				"giantswarm.io/app-build-suite-version": "2.3.0",
				"giantswarm.io/app-test-suite-status":   "unknown",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := component.New("some-app")
			if err != nil {
				t.Fatal(err)
			}

			applyBuildToolchain(c, tt.ci, pins)

			if diff := cmp.Diff(tt.wantLabels, c.Labels, cmp.Transformer("nilmap", nilToEmpty)); diff != "" {
				t.Errorf("labels mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantAnnotations, c.Annotations, cmp.Transformer("nilmap", nilToEmpty)); diff != "" {
				t.Errorf("annotations mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func nilToEmpty(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}

	return m
}
