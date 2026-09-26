package repositories

import (
	"strings"

	"go.yaml.in/yaml/v3"
)

// The architect orb, as referenced from a repo's CircleCI orbs block:
// `architect: giantswarm/architect@10.3.0`.
const (
	architectOrbPrefix = "giantswarm/architect@"

	// The alias the orb is imported under unless the repo names another.
	defaultArchitectOrbAlias = "architect"
)

// Orb job names whose invocation says which parts of the build toolchain a
// repo actually uses.
//
// push-to-app-catalog runs on the app-build-suite executor, which the repo
// cannot override, so using the job means building charts with whatever ABS
// the orb pins. run-tests-with-ats runs app-test-suite at the version the job
// parameter names, defaulting to the orb's pin. push-to-registries is read
// for its force-public parameter.
const (
	jobPushToAppCatalog = "push-to-app-catalog"
	jobRunTestsWithATS  = "run-tests-with-ats"
	jobPushToRegistries = "push-to-registries"

	atsContainerTagParameter = "app-test-suite_container_tag"
	forcePublicParameter     = "force-public"
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

// CircleCIConfigDetails is what the importer needs from a repo's CircleCI
// configuration. Everything here is declared configuration on the default
// branch: it says what a build there would use, not what the last build ran.
type CircleCIConfigDetails struct {
	// DynamicSetup reports whether config.yml is a dynamic-config setup
	// workflow, in which case the details below were merged from the files it
	// continues with (workflows.yml and, if present, custom.yml).
	DynamicSetup bool

	// Incomplete reports that part of the configuration could not be read: a
	// file was not valid YAML, a continued file could not be fetched, or a
	// setup workflow continues with something other than workflows.yml. A fact
	// missing below then means "could not tell", not "not used".
	Incomplete bool

	// ArchitectOrbRef is the raw ref after `giantswarm/architect@`, e.g.
	// `10.3.0`, `v10.3.0`, `dev:abc123` or `volatile`. Empty when the orb is
	// not referenced.
	ArchitectOrbRef string

	// UsesPushToAppCatalog reports whether the config invokes
	// push-to-app-catalog, as a workflow job or as a step of a repo-defined
	// job or command.
	UsesPushToAppCatalog bool

	// UsesRunTestsWithATS reports whether the config invokes
	// run-tests-with-ats, as a workflow job or as a step of a repo-defined
	// job or command.
	UsesRunTestsWithATS bool

	// ATSContainerTag is the app-test-suite_container_tag the repo passes to
	// run-tests-with-ats, overriding the orb's default. Empty when the repo
	// relies on the default, or when invocations disagree (see
	// ATSContainerTagConflict). May be a CircleCI template expression such as
	// `<< parameters.ats_version >>`, which callers must not publish as a
	// version.
	ATSContainerTag string

	// ATSContainerTagConflict is set when run-tests-with-ats invocations would
	// run more than one app-test-suite version: two different overrides, or
	// an override alongside an invocation on the orb default. There is then no
	// single answer to "which ATS does this repo use", and callers must not
	// pick one.
	ATSContainerTagConflict bool

	// ForcePublicRegistry reports whether any push-to-registries invocation
	// sets force-public: true.
	ForcePublicRegistry bool
}

// circleCIFile is what one CircleCI YAML file declares, before merging.
type circleCIFile struct {
	unparseable bool
	setup       bool
	orbAlias    string
	orbRef      string
	invocations []jobInvocation
}

// jobInvocation is one use of a job: a workflow job entry, or a step in a
// repo-defined job or command. params is the invocation's mapping, nil for a
// bare string entry.
type jobInvocation struct {
	name   string
	params *yaml.Node
}

// parseCircleCIFile extracts what the importer needs from one CircleCI YAML
// file. It decodes into a yaml.Node rather than a typed struct so that a key of
// an unexpected shape — the legacy `workflows: version: 2`, say — costs only
// that key instead of the whole file, and so that scalars keep their source
// text: an unquoted `1.0` stays "1.0" rather than becoming "1".
func parseCircleCIFile(text string) circleCIFile {
	file := circleCIFile{}

	var root yaml.Node
	if err := yaml.Unmarshal([]byte(text), &root); err != nil {
		file.unparseable = true

		return file
	}
	if len(root.Content) == 0 {
		return file
	}
	doc := resolve(root.Content[0])

	if setup := mappingValue(doc, "setup"); setup != nil {
		var value bool
		if setup.Decode(&value) == nil {
			file.setup = value
		}
	}

	// First reference in document order wins, should a file (nonsensically)
	// import the orb under two aliases. Inline orb definitions are mappings,
	// not references, and are skipped.
	forEachPair(mappingValue(doc, "orbs"), func(alias string, value *yaml.Node) bool {
		ref := scalarValue(value)
		if !strings.HasPrefix(ref, architectOrbPrefix) {
			return true
		}
		file.orbAlias = alias
		file.orbRef = strings.TrimPrefix(ref, architectOrbPrefix)

		return false
	})

	// Workflow jobs. Anything under workflows that is not a mapping with a
	// jobs list, such as the legacy `version: 2`, is not a workflow.
	forEachPair(mappingValue(doc, "workflows"), func(_ string, workflow *yaml.Node) bool {
		file.noteInvocations(mappingValue(workflow, "jobs"))

		return true
	})

	// Orb jobs and commands invoked as steps of the repo's own jobs and
	// commands, e.g. a `run-ats` command wrapping architect/run-tests-with-ats.
	for _, section := range []string{"jobs", "commands"} {
		forEachPair(mappingValue(doc, section), func(_ string, definition *yaml.Node) bool {
			file.noteInvocations(mappingValue(definition, "steps"))

			return true
		})
	}

	return file
}

// noteInvocations records the entries of a workflow's jobs list or a job's
// steps list. An entry is a bare name or a single-key mapping from name to
// parameters.
func (f *circleCIFile) noteInvocations(list *yaml.Node) {
	if list == nil || list.Kind != yaml.SequenceNode {
		return
	}
	for _, entry := range list.Content {
		entry = resolve(entry)
		switch entry.Kind {
		case yaml.ScalarNode:
			f.invocations = append(f.invocations, jobInvocation{name: entry.Value})
		case yaml.MappingNode:
			forEachPair(entry, func(name string, params *yaml.Node) bool {
				f.invocations = append(f.invocations, jobInvocation{name: name, params: params})

				return true
			})
		}
	}
}

// mergeCircleCIFiles combines what the files of one repo declare. The first
// file is config.yml; any others are the files a setup workflow continues
// with. The orb ref is the first one declared, jobs match under any alias a
// file imports the orb as, job usage is the union, and ATS
// overrides must agree across all invocations in all files or they count as a
// conflict.
func mergeCircleCIFiles(files ...circleCIFile) CircleCIConfigDetails {
	details := CircleCIConfigDetails{}
	aliases := make(map[string]bool)

	for i, file := range files {
		if i == 0 {
			details.DynamicSetup = file.setup
		}
		details.Incomplete = details.Incomplete || file.unparseable
		if details.ArchitectOrbRef == "" {
			details.ArchitectOrbRef = file.orbRef
		}
		if file.orbAlias != "" {
			aliases[file.orbAlias] = true
		}
	}
	if len(aliases) == 0 {
		aliases[defaultArchitectOrbAlias] = true
	}

	atsTags := make(map[string]bool)
	atsOnDefault := false

	for _, file := range files {
		for _, invocation := range file.invocations {
			switch {
			case isOrbJob(invocation.name, aliases, jobPushToAppCatalog):
				details.UsesPushToAppCatalog = true
			case isOrbJob(invocation.name, aliases, jobRunTestsWithATS):
				details.UsesRunTestsWithATS = true
				if tag := scalarValue(mappingValue(invocation.params, atsContainerTagParameter)); tag != "" {
					atsTags[tag] = true
				} else {
					atsOnDefault = true
				}
			case isOrbJob(invocation.name, aliases, jobPushToRegistries):
				var forcePublic bool
				if node := mappingValue(invocation.params, forcePublicParameter); node != nil && node.Decode(&forcePublic) == nil && forcePublic {
					details.ForcePublicRegistry = true
				}
			}
		}
	}

	switch {
	case len(atsTags) > 1, len(atsTags) == 1 && atsOnDefault:
		details.ATSContainerTagConflict = true
	case len(atsTags) == 1:
		for tag := range atsTags {
			details.ATSContainerTag = tag
		}
	}

	return details
}

// isOrbJob reports whether name invokes the given orb job: exactly the job
// name, or the name under an alias the orb is imported as in any of the
// files. Nothing looser — `other-orb/push-to-app-catalog` or
// `custom-push-to-app-catalog` is a different job.
func isOrbJob(name string, aliases map[string]bool, job string) bool {
	if name == job {
		return true
	}
	alias, rest, found := strings.Cut(name, "/")

	return found && rest == job && aliases[alias]
}

// resolve follows YAML aliases (`*anchor`) to the node they point at.
func resolve(n *yaml.Node) *yaml.Node {
	for n != nil && n.Kind == yaml.AliasNode {
		n = n.Alias
	}

	return n
}

// mappingValue returns the value under key in a mapping node, or nil. Keys
// brought in by a YAML merge key (`<<: *defaults`) count, with keys written
// directly taking precedence, as in YAML itself.
func mappingValue(n *yaml.Node, key string) *yaml.Node {
	n = resolve(n)
	if n == nil || n.Kind != yaml.MappingNode {
		return nil
	}

	var merged []*yaml.Node
	for i := 0; i+1 < len(n.Content); i += 2 {
		switch n.Content[i].Value {
		case key:
			return resolve(n.Content[i+1])
		case "<<":
			merged = append(merged, n.Content[i+1])
		}
	}

	for _, m := range merged {
		m = resolve(m)
		sources := []*yaml.Node{m}
		if m.Kind == yaml.SequenceNode {
			sources = m.Content
		}
		for _, source := range sources {
			if value := mappingValue(source, key); value != nil {
				return value
			}
		}
	}

	return nil
}

// forEachPair calls fn for every key/value pair of a mapping node in document
// order, until fn returns false. Anything but a mapping is ignored.
func forEachPair(n *yaml.Node, fn func(key string, value *yaml.Node) bool) {
	n = resolve(n)
	if n == nil || n.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if !fn(n.Content[i].Value, resolve(n.Content[i+1])) {
			return
		}
	}
}

// scalarValue returns a scalar's source text, as a CircleCI parameter would
// see it: `0.4`, `1.0` and `"1.0"` all name a container tag verbatim. Nulls,
// booleans and non-scalars yield "".
func scalarValue(n *yaml.Node) string {
	n = resolve(n)
	if n == nil || n.Kind != yaml.ScalarNode {
		return ""
	}
	switch n.ShortTag() {
	case "!!null", "!!bool":
		return ""
	}

	return strings.TrimSpace(n.Value)
}
