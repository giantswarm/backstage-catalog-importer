package repositories

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseCircleCIFile(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   CircleCIConfigDetails
	}{
		{
			name: "release orb, chart push and ATS with orb default",
			config: `version: 2.1
orbs:
  architect: giantswarm/architect@10.3.0
workflows:
  build:
    jobs:
      - architect/push-to-app-catalog:
          context: architect
          name: push-to-catalog
      - architect/run-tests-with-ats:
          name: execute chart tests
          requires:
            - push-to-catalog
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:      "10.3.0",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
			},
		},
		{
			name: "v-prefixed orb ref is kept raw",
			config: `orbs:
  architect: giantswarm/architect@v10.3.0
`,
			want: CircleCIConfigDetails{ArchitectOrbRef: "v10.3.0"},
		},
		{
			name: "dev orb ref",
			config: `orbs:
  architect: giantswarm/architect@dev:abc123
`,
			want: CircleCIConfigDetails{ArchitectOrbRef: "dev:abc123"},
		},
		{
			name: "volatile orb ref",
			config: `orbs:
  architect: giantswarm/architect@volatile
`,
			want: CircleCIConfigDetails{ArchitectOrbRef: "volatile"},
		},
		{
			name: "orb referenced under a different alias",
			config: `orbs:
  gs: giantswarm/architect@9.6.0
workflows:
  build:
    jobs:
      - gs/push-to-app-catalog:
          context: architect
`,
			want: CircleCIConfigDetails{ArchitectOrbRef: "9.6.0", UsesPushToAppCatalog: true},
		},
		{
			name: "architect orb not used, other orbs and inline orb ignored",
			config: `orbs:
  slack: circleci/slack@4.12.5
  local:
    jobs:
      noop:
        docker:
          - image: alpine
        steps:
          - run: "true"
workflows:
  build:
    jobs:
      - local/noop
`,
			want: CircleCIConfigDetails{},
		},
		{
			name: "jobs without orb prefix, given as bare strings",
			config: `orbs:
  architect: giantswarm/architect@8.0.0
workflows:
  build:
    jobs:
      - push-to-app-catalog
      - run-tests-with-ats
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:      "8.0.0",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
			},
		},
		{
			name: "ATS container tag override",
			config: `orbs:
  architect: giantswarm/architect@10.1.2
workflows:
  build:
    jobs:
      - architect/run-tests-with-ats:
          name: execute chart tests
          app-test-suite_version: v0.4.1
          app-test-suite_container_tag: 0.4.1
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:     "10.1.2",
				UsesRunTestsWithATS: true,
				ATSContainerTag:     "0.4.1",
			},
		},
		{
			name: "ATS container tag override given as a float scalar",
			config: `orbs:
  architect: giantswarm/architect@10.1.2
workflows:
  build:
    jobs:
      - architect/run-tests-with-ats:
          app-test-suite_container_tag: 0.4
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:     "10.1.2",
				UsesRunTestsWithATS: true,
				ATSContainerTag:     "0.4",
			},
		},
		{
			name: "same ATS override in two jobs is one answer",
			config: `orbs:
  architect: giantswarm/architect@10.1.2
workflows:
  build:
    jobs:
      - architect/run-tests-with-ats:
          name: smoke
          app-test-suite_container_tag: "0.10.6"
      - architect/run-tests-with-ats:
          name: functional
          app-test-suite_container_tag: "0.10.6"
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:     "10.1.2",
				UsesRunTestsWithATS: true,
				ATSContainerTag:     "0.10.6",
			},
		},
		{
			name: "conflicting ATS overrides give no answer",
			config: `orbs:
  architect: giantswarm/architect@10.1.2
workflows:
  build:
    jobs:
      - architect/run-tests-with-ats:
          name: old
          app-test-suite_container_tag: "0.10.6"
      - architect/run-tests-with-ats:
          name: new
          app-test-suite_container_tag: "0.15.0"
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:         "10.1.2",
				UsesRunTestsWithATS:     true,
				ATSContainerTagConflict: true,
			},
		},
		{
			name: "only the exact orb job name matches, not look-alikes",
			config: `orbs:
  architect: giantswarm/architect@10.1.2
  other: someone/other@1.0.0
workflows:
  build:
    jobs:
      - architect/push-to-app-catalog-dryrun
      - other/push-to-app-catalog
      - custom-push-to-app-catalog
      - other/run-tests-with-ats:
          app-test-suite_container_tag: "0.1.0"
      - architect/go-build
`,
			want: CircleCIConfigDetails{ArchitectOrbRef: "10.1.2"},
		},
		{
			name: "jobs under the default alias do not match when the orb is imported under another",
			config: `orbs:
  gs: giantswarm/architect@9.6.0
workflows:
  build:
    jobs:
      - architect/push-to-app-catalog
`,
			want: CircleCIConfigDetails{ArchitectOrbRef: "9.6.0"},
		},
		{
			name: "legacy workflows version key does not discard the rest of the file",
			config: `version: 2.1
orbs:
  architect: giantswarm/architect@10.1.2
workflows:
  version: 2
  build:
    jobs:
      - architect/push-to-app-catalog:
          name: push-to-default-app-catalog
      - architect/run-tests-with-ats:
          app-test-suite_container_tag: 0.10.6
      - architect/push-to-registries:
          force-public: true
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:      "10.1.2",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
				ATSContainerTag:      "0.10.6",
				ForcePublicRegistry:  true,
			},
		},
		{
			name: "an ATS override alongside a job on the orb default is a conflict",
			config: `orbs:
  architect: giantswarm/architect@10.1.2
workflows:
  build:
    jobs:
      - architect/run-tests-with-ats:
          name: pinned
          app-test-suite_container_tag: "0.10.6"
      - architect/run-tests-with-ats:
          name: default
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:         "10.1.2",
				UsesRunTestsWithATS:     true,
				ATSContainerTagConflict: true,
			},
		},
		{
			name: "unquoted float ATS tag keeps its trailing zero",
			config: `orbs:
  architect: giantswarm/architect@10.1.2
workflows:
  build:
    jobs:
      - architect/run-tests-with-ats:
          app-test-suite_container_tag: 1.0
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:     "10.1.2",
				UsesRunTestsWithATS: true,
				ATSContainerTag:     "1.0",
			},
		},
		{
			name: "orb job invoked as a step of a repo-defined command",
			config: `version: 2.1
orbs:
  architect: giantswarm/architect@10.6.2
commands:
  run-ats:
    parameters:
      ats_version:
        type: string
        default: 0.10.2
    steps:
    - architect/run-tests-with-ats:
        app-test-suite_container_tag: << parameters.ats_version >>
jobs:
  run-tests:
    machine:
      image: ubuntu-2204:current
    steps:
      - checkout
      - run-ats
workflows:
  build:
    jobs:
    - architect/push-to-app-catalog:
        name: push-kyverno-chart-to-giantswarm-catalog
    - run-tests
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:      "10.6.2",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
				ATSContainerTag:      "<< parameters.ats_version >>",
			},
		},
		{
			name: "parameters brought in through a YAML anchor and merge key",
			config: `orbs:
  architect: giantswarm/architect@10.1.2
ats_defaults: &ats_defaults
  app-test-suite_container_tag: "0.10.6"
workflows:
  build:
    jobs:
      - architect/run-tests-with-ats:
          <<: *ats_defaults
          name: e2e
`,
			want: CircleCIConfigDetails{
				ArchitectOrbRef:     "10.1.2",
				UsesRunTestsWithATS: true,
				ATSContainerTag:     "0.10.6",
			},
		},
		{
			name: "dynamic-config setup workflow declares nothing itself",
			config: `version: 2.1
setup: true
orbs:
  continuation: circleci/continuation@2.0.1
jobs:
  setup:
    executor: continuation/default
    steps:
      - checkout
      - continuation/continue:
          configuration_path: .circleci/continue.yml
workflows:
  setup:
    jobs:
      - setup
`,
			want: CircleCIConfigDetails{DynamicSetup: true},
		},
		{
			name:   "empty config",
			config: "",
			want:   CircleCIConfigDetails{},
		},
		{
			name:   "invalid YAML is incomplete, not empty",
			config: "orbs: [unclosed",
			want:   CircleCIConfigDetails{Incomplete: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeCircleCIFiles(parseCircleCIFile(tt.config))
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mergeCircleCIFiles(parseCircleCIFile()) mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMergeCircleCIFiles(t *testing.T) {
	setup := parseCircleCIFile(`version: 2.1
setup: true
orbs:
  continuation: circleci/continuation@2.0.1
`)
	workflows := parseCircleCIFile(`version: 2.1
orbs:
  architect: giantswarm/architect@10.1.0
workflows:
  build:
    jobs:
      - architect/go-build:
          name: go-build
      - architect/push-to-app-catalog:
          name: push-to-catalog
`)
	customWithATS := parseCircleCIFile(`workflows:
  build:
    jobs:
      - architect/run-tests-with-ats:
          name: e2e
          app-test-suite_container_tag: "0.10.6"
`)
	customWithOtherATS := parseCircleCIFile(`workflows:
  build:
    jobs:
      - architect/run-tests-with-ats:
          name: e2e-new
          app-test-suite_container_tag: "0.15.0"
`)

	tests := []struct {
		name  string
		files []circleCIFile
		want  CircleCIConfigDetails
	}{
		{
			name:  "setup plus generated workflows",
			files: []circleCIFile{setup, workflows},
			want: CircleCIConfigDetails{
				DynamicSetup:         true,
				ArchitectOrbRef:      "10.1.0",
				UsesPushToAppCatalog: true,
			},
		},
		{
			name:  "custom.yml adds an ATS job with an override",
			files: []circleCIFile{setup, workflows, customWithATS},
			want: CircleCIConfigDetails{
				DynamicSetup:         true,
				ArchitectOrbRef:      "10.1.0",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
				ATSContainerTag:      "0.10.6",
			},
		},
		{
			name:  "overrides disagreeing across files are a conflict",
			files: []circleCIFile{setup, workflows, customWithATS, customWithOtherATS},
			want: CircleCIConfigDetails{
				DynamicSetup:            true,
				ArchitectOrbRef:         "10.1.0",
				UsesPushToAppCatalog:    true,
				UsesRunTestsWithATS:     true,
				ATSContainerTagConflict: true,
			},
		},
		{
			name: "orb imported under different aliases in different files",
			files: []circleCIFile{setup, workflows, parseCircleCIFile(`orbs:
  gs: giantswarm/architect@10.1.0
workflows:
  test:
    jobs:
      - gs/run-tests-with-ats
`)},
			want: CircleCIConfigDetails{
				DynamicSetup:         true,
				ArchitectOrbRef:      "10.1.0",
				UsesPushToAppCatalog: true,
				UsesRunTestsWithATS:  true,
			},
		},
		{
			name:  "setup workflow whose continuation could not be read",
			files: []circleCIFile{setup},
			want:  CircleCIConfigDetails{DynamicSetup: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeCircleCIFiles(tt.files...)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("mergeCircleCIFiles() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
