// Package architectorb resolves what a given release of the giantswarm/architect
// CircleCI orb pins: the app-build-suite image its executor runs, and the
// app-test-suite container tag its run-tests-with-ats job defaults to.
//
// Repos do not choose these directly. A repo names an orb version in its
// .circleci/config.yml, and the orb decides the rest, so "which ABS builds this
// chart" is a question about the orb's source at that tag.
package architectorb

import (
	"context"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/google/go-github/v92/github"
	"go.yaml.in/yaml/v3"
)

const (
	// Owner and repository of the orb source.
	Owner      = "giantswarm"
	Repository = "architect-orb"

	// Where the orb pins live. The orb has used these names since v3; the
	// resolver still falls back to a .yml spelling in case that changes.
	executorPath = "src/executors/app-build-suite"
	atsJobPath   = "src/jobs/run-tests-with-ats"

	// The ABS image is tagged for CircleCI use. The suffix is packaging, not
	// version, and is dropped.
	absImageSuffix = "-circleci"

	atsContainerTagParameter = "app-test-suite_container_tag"
)

var (
	// releaseVersion is the shape of an orb release ref, with or without a
	// leading v. Anything else (dev:<sha>, volatile, a branch name) is not a
	// release and has no tag in the orb repository to look up.
	releaseVersion = regexp.MustCompile(`^v?(\d+\.\d+\.\d+)$`)

	yamlExtensions = []string{".yaml", ".yml"}
)

// Pins is what an orb release fixes for every repo that uses it. An empty
// field means unknown: the file was unreadable at that tag or carried no
// default. It never means "none".
type Pins struct {
	// AppBuildSuite is the ABS version the push-to-app-catalog executor runs,
	// e.g. "2.3.0".
	AppBuildSuite string

	// AppTestSuite is the ATS container tag run-tests-with-ats defaults to,
	// e.g. "0.15.0". Empty for orb releases predating the job (before v5) or
	// where the parameter has no default.
	AppTestSuite string
}

// ReleaseVersion normalizes an orb ref to a bare release version ("10.3.0")
// and reports whether the ref is a release at all.
func ReleaseVersion(ref string) (string, bool) {
	match := releaseVersion.FindStringSubmatch(strings.TrimSpace(ref))
	if match == nil {
		return "", false
	}

	return match[1], true
}

// ContentGetter is the slice of the go-github repositories API the resolver
// needs. *github.RepositoriesService satisfies it.
type ContentGetter interface {
	GetContents(ctx context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error)
}

// Resolver looks up orb pins by release version, once per version per run.
type Resolver struct {
	ctx    context.Context
	getter ContentGetter
	mu     sync.Mutex
	cache  map[string]Pins
}

// NewResolver returns a resolver backed by the given GitHub content API.
func NewResolver(ctx context.Context, getter ContentGetter) *Resolver {
	return &Resolver{
		ctx:    ctx,
		getter: getter,
		cache:  make(map[string]Pins),
	}
}

// Pins returns what the orb release pins. Definite results are cached for the
// lifetime of the resolver, negative ones included, so a fleet on a dozen orb
// versions costs a dozen pairs of requests, not one per repo. A result that
// rests on a failed request (transport error, 5xx, a 403 rate limit) is
// returned but not cached, and the next repo on that version asks again:
// otherwise one unlucky request would cost every repo on the orb release its
// labels for the rest of the run.
func (r *Resolver) Pins(version string) Pins {
	r.mu.Lock()
	defer r.mu.Unlock()

	if pins, ok := r.cache[version]; ok {
		return pins
	}

	ref := "v" + version
	pins := Pins{}

	executor, executorDefinite := r.load(ref, executorPath)
	pins.AppBuildSuite = ParseExecutorImageTag(executor)

	atsJob, atsJobDefinite := r.load(ref, atsJobPath)
	pins.AppTestSuite = ParseATSDefaultContainerTag(atsJob)

	if pins.AppBuildSuite == "" || pins.AppTestSuite == "" {
		log.Printf("DEBUG - architect-orb %s - pins partially unknown: abs=%q ats=%q\n", ref, pins.AppBuildSuite, pins.AppTestSuite)
	}

	if executorDefinite && atsJobDefinite {
		r.cache[version] = pins
	}

	return pins
}

// load fetches a source file at ref, trying each YAML extension in turn. It
// returns the content ("" when there is none) and whether that answer is
// definite: the file was read, or every spelling of it is a 404. Any other
// failure is not.
func (r *Resolver) load(ref, basePath string) (string, bool) {
	for _, ext := range yamlExtensions {
		path := basePath + ext
		file, _, resp, err := r.getter.GetContents(r.ctx, Owner, Repository, path, &github.RepositoryContentGetOptions{Ref: ref})
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				continue
			}
			log.Printf("WARN - architect-orb %s - could not read %s: %v\n", ref, path, err)

			return "", false
		}
		if file == nil {
			continue
		}
		content, err := file.GetContent()
		if err != nil {
			log.Printf("WARN - architect-orb %s - could not decode %s: %v\n", ref, path, err)

			return "", false
		}

		return content, true
	}

	return "", true
}

// ParseExecutorImageTag reads the ABS version out of the orb's
// app-build-suite executor definition. The executor's first docker image is
// the ABS image; its tag, minus the -circleci packaging suffix, is the
// version. Returns "" when no such image is declared.
func ParseExecutorImageTag(executorYAML string) string {
	var executor struct {
		Docker []struct {
			Image string `yaml:"image"`
		} `yaml:"docker"`
	}

	if err := yaml.Unmarshal([]byte(executorYAML), &executor); err != nil {
		return ""
	}

	for _, container := range executor.Docker {
		image := strings.TrimSpace(container.Image)
		if !strings.Contains(image, "app-build-suite") {
			continue
		}
		// Split on the last colon so a registry port would not confuse it.
		idx := strings.LastIndex(image, ":")
		if idx < 0 || strings.Contains(image[idx:], "/") {
			return ""
		}
		tag := strings.TrimSuffix(image[idx+1:], absImageSuffix)
		if tag == "" {
			return ""
		}

		return tag
	}

	return ""
}

// ParseATSDefaultContainerTag reads the default app-test-suite container tag
// out of the orb's run-tests-with-ats job definition. Returns "" when the
// parameter is absent or has no default. The default is taken as written: an
// unquoted `1.0` is the tag "1.0", not the float 1.
func ParseATSDefaultContainerTag(jobYAML string) string {
	var job struct {
		Parameters map[string]struct {
			Default yaml.Node `yaml:"default"`
		} `yaml:"parameters"`
	}

	if err := yaml.Unmarshal([]byte(jobYAML), &job); err != nil {
		return ""
	}

	param, ok := job.Parameters[atsContainerTagParameter]
	if !ok || param.Default.Kind != yaml.ScalarNode {
		return ""
	}
	switch param.Default.ShortTag() {
	case "!!null", "!!bool":
		return ""
	}

	return strings.TrimSpace(param.Default.Value)
}
