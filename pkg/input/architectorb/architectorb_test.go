package architectorb

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v91/github"
)

// executorV10 is src/executors/app-build-suite.yaml at architect-orb v10.3.0.
const executorV10 = `docker:
  - entrypoint: /bin/bash
    # registry: gsoci.azurecr.io/giantswarm/app-build-suite
    image: gsoci.azurecr.io/giantswarm/app-build-suite:2.3.0-circleci
`

// atsJobV10 is the parameters block of src/jobs/run-tests-with-ats.yaml at
// architect-orb v10.3.0.
const atsJobV10 = `parameters:
  chart_archive_prefix:
    description: "Prefix for the chart archive file to execute tests for."
    type: string
    default: ""
  app-test-suite_version:
    description: "Version of app-test-suite dabs.sh container wrapper to use (git tag or commit)"
    type: string
    default: "v0.15.0"
  app-test-suite_container_tag:
    description: "Container tag of app-test-suite to use (check gsoci.azurecr.io/giantswarm/app-test-suite)"
    type: string
    default: "0.15.0"
  additional_app-test-suite_flags:
    description: "Additional app-test-suite flags to use"
    type: string
    default: ""
`

func TestReleaseVersion(t *testing.T) {
	tests := []struct {
		ref     string
		want    string
		release bool
	}{
		{"10.3.0", "10.3.0", true},
		{"v10.3.0", "10.3.0", true},
		{" 9.6.0 ", "9.6.0", true},
		{"dev:abc123", "", false},
		{"volatile", "", false},
		{"10.3", "", false},
		{"10.3.0-rc.1", "", false},
		{"main", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		got, release := ReleaseVersion(tt.ref)
		if got != tt.want || release != tt.release {
			t.Errorf("ReleaseVersion(%q) = (%q, %v), want (%q, %v)", tt.ref, got, release, tt.want, tt.release)
		}
	}
}

func TestParseExecutorImageTag(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want string
	}{
		{"v10 executor", executorV10, "2.3.0"},
		{
			name: "older registry, no circleci suffix",
			yaml: `docker:
  - image: quay.io/giantswarm/app-build-suite:1.1.4
`,
			want: "1.1.4",
		},
		{
			name: "unrelated image only",
			yaml: `docker:
  - image: cimg/go:1.22
`,
			want: "",
		},
		{
			name: "image without tag",
			yaml: `docker:
  - image: gsoci.azurecr.io/giantswarm/app-build-suite
`,
			want: "",
		},
		{"empty", "", ""},
		{"invalid", "docker: [", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseExecutorImageTag(tt.yaml); got != tt.want {
				t.Errorf("ParseExecutorImageTag() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseATSDefaultContainerTag(t *testing.T) {
	tests := []struct {
		name string
		yaml string
		want string
	}{
		{"v10 job", atsJobV10, "0.15.0"},
		{
			name: "parameter without default",
			yaml: `parameters:
  app-test-suite_container_tag:
    description: "Container tag of app-test-suite to use"
    type: string
`,
			want: "",
		},
		{
			name: "float default",
			yaml: `parameters:
  app-test-suite_container_tag:
    type: string
    default: 0.4
`,
			want: "0.4",
		},
		{
			name: "parameter absent",
			yaml: `parameters:
  chart_archive_prefix:
    type: string
    default: ""
`,
			want: "",
		},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseATSDefaultContainerTag(tt.yaml); got != tt.want {
				t.Errorf("ParseATSDefaultContainerTag() = %q, want %q", got, tt.want)
			}
		})
	}
}

// fakeGetter serves orb source files by "<ref>:<path>", records the requests
// it saw, and answers 404 for anything else.
type fakeGetter struct {
	files    map[string]string
	requests []string
	failWith error
}

func (f *fakeGetter) GetContents(_ context.Context, owner, repo, path string, opts *github.RepositoryContentGetOptions) (*github.RepositoryContent, []*github.RepositoryContent, *github.Response, error) {
	key := opts.Ref + ":" + path
	f.requests = append(f.requests, owner+"/"+repo+"@"+key)

	if f.failWith != nil {
		return nil, nil, &github.Response{Response: &http.Response{StatusCode: http.StatusBadGateway}}, f.failWith
	}

	content, ok := f.files[key]
	if !ok {
		return nil, nil, &github.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, errors.New("not found")
	}

	encoding := ""

	return &github.RepositoryContent{Content: &content, Encoding: &encoding}, nil, &github.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
}

func newResolver(getter *fakeGetter) *Resolver {
	r := NewResolver(context.Background(), getter)
	r.logging = false

	return r
}

func TestResolver_Pins(t *testing.T) {
	getter := &fakeGetter{files: map[string]string{
		"v10.3.0:src/executors/app-build-suite.yaml": executorV10,
		"v10.3.0:src/jobs/run-tests-with-ats.yaml":   atsJobV10,
	}}
	r := newResolver(getter)

	got := r.Pins("10.3.0")
	want := Pins{AppBuildSuite: "2.3.0", AppTestSuite: "0.15.0"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Pins() mismatch (-want +got):\n%s", diff)
	}

	// Second call for the same version is served from cache.
	requests := len(getter.requests)
	if again := r.Pins("10.3.0"); again != got {
		t.Errorf("cached Pins() = %+v, want %+v", again, got)
	}
	if len(getter.requests) != requests {
		t.Errorf("cached Pins() issued %d more requests", len(getter.requests)-requests)
	}
	if getter.requests[0] != "giantswarm/architect-orb@v10.3.0:src/executors/app-build-suite.yaml" {
		t.Errorf("first request = %q, want the executor at tag v10.3.0", getter.requests[0])
	}
}

func TestResolver_Pins_FallsBackToYml(t *testing.T) {
	getter := &fakeGetter{files: map[string]string{
		"v4.0.0:src/executors/app-build-suite.yml": `docker:
  - image: quay.io/giantswarm/app-build-suite:0.2.4
`,
		// No run-tests-with-ats job at all: the orb predates it.
	}}
	r := newResolver(getter)

	got := r.Pins("4.0.0")
	want := Pins{AppBuildSuite: "0.2.4"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("Pins() mismatch (-want +got):\n%s", diff)
	}
}

func TestResolver_Pins_UnknownStaysUnknown(t *testing.T) {
	t.Run("tag does not exist", func(t *testing.T) {
		r := newResolver(&fakeGetter{files: map[string]string{}})
		if got := r.Pins("99.0.0"); got != (Pins{}) {
			t.Errorf("Pins() = %+v, want empty", got)
		}
	})

	t.Run("API error", func(t *testing.T) {
		getter := &fakeGetter{failWith: errors.New("bad gateway")}
		r := newResolver(getter)
		if got := r.Pins("10.3.0"); got != (Pins{}) {
			t.Errorf("Pins() = %+v, want empty", got)
		}
		// A non-404 error is not retried against the other extension, and the
		// negative result is cached like any other.
		if len(getter.requests) != 2 {
			t.Errorf("issued %d requests, want 2 (one per file)", len(getter.requests))
		}
		r.Pins("10.3.0")
		if len(getter.requests) != 2 {
			t.Errorf("negative result was not cached: %d requests", len(getter.requests))
		}
	})
}
