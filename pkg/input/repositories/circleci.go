package repositories

import (
	"fmt"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// The architect orb, as referenced from a repo's CircleCI orbs block:
// `architect: giantswarm/architect@10.3.0`.
const architectOrbPrefix = "giantswarm/architect@"

// Job names (with or without the `architect/` orb prefix) whose presence in a
// workflow says which parts of the build toolchain a repo actually uses.
//
// push-to-app-catalog runs on the app-build-suite executor, which the repo
// cannot override, so using the job means building charts with whatever ABS
// the orb pins. run-tests-with-ats runs app-test-suite at the version the job
// parameter names, defaulting to the orb's pin.
const (
	jobPushToAppCatalog = "push-to-app-catalog"
	jobRunTestsWithATS  = "run-tests-with-ats"

	atsContainerTagParameter = "app-test-suite_container_tag"
)

// CircleCI config files. config.yml is always read. A devctl-generated config
// is a dynamic-config setup workflow (`setup: true`) that continues with the
// generated workflows.yml, deep-merged with the optional repo-owned custom.yml
// — so for those repos the orb reference and the jobs live in the other two
// files, and all three have to be read to see what actually runs.
const (
	circleCIConfigPath    = ".circleci/config.yml"
	circleCIWorkflowsPath = ".circleci/workflows.yml"
	circleCICustomPath    = ".circleci/custom.yml"
)

// CircleCIConfigDetails is what the build toolchain view needs from a repo's
// CircleCI configuration. Everything here is declared configuration on the
// default branch: it says what a build there would use, not what the last
// build ran.
type CircleCIConfigDetails struct {
	// DynamicSetup reports whether config.yml is a dynamic-config setup
	// workflow, in which case the details below were merged from the files it
	// continues with (workflows.yml and, if present, custom.yml).
	DynamicSetup bool

	// ArchitectOrbRef is the raw ref after `giantswarm/architect@`, e.g.
	// `10.3.0`, `v10.3.0`, `dev:abc123` or `volatile`. Empty when the orb is
	// not referenced.
	ArchitectOrbRef string

	// UsesPushToAppCatalog reports whether any workflow runs a job named
	// push-to-app-catalog (with or without the architect/ prefix).
	UsesPushToAppCatalog bool

	// UsesRunTestsWithATS reports whether any workflow runs a job named
	// run-tests-with-ats (with or without the architect/ prefix).
	UsesRunTestsWithATS bool

	// ATSContainerTag is the app-test-suite_container_tag the repo passes to
	// run-tests-with-ats, overriding the orb's default. Empty when the repo
	// relies on the default, or when several jobs disagree (see
	// ATSContainerTagConflict).
	ATSContainerTag string

	// ATSContainerTagConflict is set when more than one run-tests-with-ats job
	// passes a container tag and they differ. There is then no single answer to
	// "which ATS does this repo use", and callers must emit nothing rather than
	// pick one.
	ATSContainerTagConflict bool
}

// circleCIFile is what one CircleCI YAML file declares, before merging.
type circleCIFile struct {
	setup            bool
	orbRef           string
	pushToAppCatalog bool
	runTestsWithATS  bool
	atsTags          map[string]bool
}

// parseCircleCIConfig reads a single, self-contained config.yml. For a
// dynamic-config setup workflow it reports DynamicSetup and nothing else; the
// loader then reads the continued files and merges them with
// mergeCircleCIFiles.
func parseCircleCIConfig(configYAML string) CircleCIConfigDetails {
	return mergeCircleCIFiles(parseCircleCIFile(configYAML))
}

// parseCircleCIFile extracts the build toolchain facts from one CircleCI YAML
// file. It is deliberately lenient about shape: orbs may be inline definitions
// rather than references, workflow job entries may be bare strings or
// single-key maps, and parameter values may be YAML scalars of any type.
// Anything unparseable yields the zero value, never an invented fact.
func parseCircleCIFile(text string) circleCIFile {
	var config struct {
		Setup     bool           `yaml:"setup"`
		Orbs      map[string]any `yaml:"orbs"`
		Workflows map[string]struct {
			Jobs []any `yaml:"jobs"`
		} `yaml:"workflows"`
	}

	file := circleCIFile{atsTags: make(map[string]bool)}

	if err := yaml.Unmarshal([]byte(text), &config); err != nil {
		return file
	}

	file.setup = config.Setup

	// Map iteration order is random; pick deterministically should a file
	// (nonsensically) reference the orb under two aliases.
	aliases := make([]string, 0, len(config.Orbs))
	for alias := range config.Orbs {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	for _, alias := range aliases {
		ref, ok := config.Orbs[alias].(string)
		if !ok {
			// An inline orb definition, not a reference.
			continue
		}
		if strings.HasPrefix(ref, architectOrbPrefix) {
			file.orbRef = strings.TrimPrefix(ref, architectOrbPrefix)

			break
		}
	}

	for _, workflow := range config.Workflows {
		for _, entry := range workflow.Jobs {
			switch job := entry.(type) {
			case string:
				file.noteJob(job)
			case map[string]any:
				for jobName, jobConfig := range job {
					file.noteJob(jobName)
					if !strings.HasSuffix(jobName, jobRunTestsWithATS) {
						continue
					}
					params, ok := jobConfig.(map[string]any)
					if !ok {
						continue
					}
					if tag := scalarString(params[atsContainerTagParameter]); tag != "" {
						file.atsTags[tag] = true
					}
				}
			}
		}
	}

	return file
}

func (f *circleCIFile) noteJob(jobName string) {
	if strings.HasSuffix(jobName, jobPushToAppCatalog) {
		f.pushToAppCatalog = true
	}
	if strings.HasSuffix(jobName, jobRunTestsWithATS) {
		f.runTestsWithATS = true
	}
}

// mergeCircleCIFiles combines what the files of one repo declare into the
// details the toolchain view needs. The first file is config.yml; any others
// are the files a setup workflow continues with. Job usage is the union, the
// orb ref is the first one declared, and ATS overrides must agree across all
// files or they count as a conflict.
func mergeCircleCIFiles(files ...circleCIFile) CircleCIConfigDetails {
	details := CircleCIConfigDetails{}
	atsTags := make(map[string]bool)

	for i, file := range files {
		if i == 0 {
			details.DynamicSetup = file.setup
		}
		if details.ArchitectOrbRef == "" {
			details.ArchitectOrbRef = file.orbRef
		}
		details.UsesPushToAppCatalog = details.UsesPushToAppCatalog || file.pushToAppCatalog
		details.UsesRunTestsWithATS = details.UsesRunTestsWithATS || file.runTestsWithATS
		for tag := range file.atsTags {
			atsTags[tag] = true
		}
	}

	switch len(atsTags) {
	case 0:
		// Relies on the orb default.
	case 1:
		for tag := range atsTags {
			details.ATSContainerTag = tag
		}
	default:
		details.ATSContainerTagConflict = true
	}

	return details
}

// scalarString renders a YAML scalar as the string a CircleCI parameter would
// see. A tag like `0.4.1` parses as a string, but `0.4` parses as a float, and
// both name a container tag.
func scalarString(v any) string {
	switch value := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(value)
	case bool:
		return ""
	default:
		return fmt.Sprint(value)
	}
}
