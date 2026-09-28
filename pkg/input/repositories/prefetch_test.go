package repositories

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-github/v92/github"
)

// fakeContentsServer answers the GitHub contents API: every repo has a
// README.md, and a repo named "charted" also has helm/app with a Chart.yaml
// and a values schema. Repos named in failing return 500 for their CircleCI
// config until healed; the first request for a repo named in rateLimitOnce is
// refused with a secondary rate limit.
type fakeContentsServer struct {
	mu            sync.Mutex
	requests      map[string]int
	failing       map[string]bool
	rateLimitOnce map[string]bool

	inFlight    atomic.Int32
	maxInFlight atomic.Int32
}

func (f *fakeContentsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	n := f.inFlight.Add(1)
	defer f.inFlight.Add(-1)
	for {
		peak := f.maxInFlight.Load()
		if n <= peak || f.maxInFlight.CompareAndSwap(peak, n) {
			break
		}
	}
	// Hold the request long enough for concurrent ones to overlap.
	time.Sleep(2 * time.Millisecond)

	// /repos/<org>/<repo>/contents/<path>
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/repos/"), "/", 4)
	if len(parts) < 4 {
		http.NotFound(w, r)

		return
	}
	repo, path := parts[1], parts[3]

	f.mu.Lock()
	f.requests[repo]++
	failing := f.failing[repo]
	rateLimited := f.rateLimitOnce[repo]
	delete(f.rateLimitOnce, repo)
	f.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	switch {
	case rateLimited:
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusForbidden)
		_, _ = fmt.Fprint(w, `{"message":"You have exceeded a secondary rate limit.","documentation_url":"https://docs.github.com/rest/overview/rate-limits-for-the-rest-api#about-secondary-rate-limits"}`)
	case failing && path == ".circleci/config.yml":
		http.Error(w, `{"message":"boom"}`, http.StatusInternalServerError)
	case path == "README.md":
		_, _ = fmt.Fprintf(w, `{"type":"file","name":"README.md","path":"README.md","content":"","encoding":"base64"}`)
	case repo == "charted" && path == "helm":
		_, _ = fmt.Fprint(w, `[{"type":"dir","name":"app","path":"helm/app"}]`)
	case repo == "charted" && path == "helm/app":
		_, _ = fmt.Fprint(w, `[{"type":"file","name":"Chart.yaml","path":"helm/app/Chart.yaml"},{"type":"file","name":"values.schema.json","path":"helm/app/values.schema.json"}]`)
	case repo == "charted" && path == "helm/app/Chart.yaml":
		_, _ = fmt.Fprintf(w, `{"type":"file","name":"Chart.yaml","path":"helm/app/Chart.yaml","content":%q,"encoding":"base64"}`, base64.StdEncoding.EncodeToString([]byte(testChartYAML)))
	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"message":"Not Found"}`)
	}
}

const testChartYAML = "apiVersion: v2\nname: app\nversion: 1.0.0\n"

func (f *fakeContentsServer) requestsFor(repo string) int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.requests[repo]
}

func newFakeService(t *testing.T, fake *fakeContentsServer) *Service {
	t.Helper()

	srv := httptest.NewServer(fake)
	t.Cleanup(srv.Close)

	baseURL := srv.URL + "/"
	// go-github's own rate limit check stays on, as in production: after a
	// rate-limited response it refuses later requests without sending them,
	// which is what PrefetchContentDetails has to cope with.
	client, err := github.NewClient(
		github.WithHTTPClient(srv.Client()),
		github.WithURLs(&baseURL, &baseURL),
	)
	if err != nil {
		t.Fatal(err)
	}

	return &Service{
		config:                   Config{GithubOrganization: "giantswarm"},
		ctx:                      context.Background(),
		githubClient:             client,
		githubRepoDetails:        make(map[string]GithubRepoDetails),
		githubRepoContentDetails: make(map[string]GithubRepoContentDetails),
	}
}

func TestPrefetchContentDetails(t *testing.T) {
	fake := &fakeContentsServer{
		requests: make(map[string]int),
		failing:  map[string]bool{"broken": true},
	}
	s := newFakeService(t, fake)

	names := []string{"broken"}
	for i := range 40 {
		names = append(names, fmt.Sprintf("repo-%02d", i))
	}

	const workers = 4
	s.PrefetchContentDetails(names, workers)

	if peak := fake.maxInFlight.Load(); peak > workers {
		t.Errorf("peak concurrent requests = %d, want at most %d", peak, workers)
	}
	if peak := fake.maxInFlight.Load(); peak < 2 {
		t.Errorf("peak concurrent requests = %d, want the loads to overlap", peak)
	}

	// Prefetched repos are served from cache.
	before := fake.requestsFor("repo-07")
	hasReadme, err := s.GetHasReadme("repo-07")
	if err != nil || !hasReadme {
		t.Errorf("GetHasReadme() = %v, %v; want true, nil", hasReadme, err)
	}
	if after := fake.requestsFor("repo-07"); after != before {
		t.Errorf("GetHasReadme() on a prefetched repo issued %d requests", after-before)
	}

	// A repo that failed is not cached as empty: the getter tries again and
	// reports the error as it would have without the prefetch...
	if _, err := s.GetHasReadme("broken"); err == nil {
		t.Error("GetHasReadme() on a repo that keeps failing returned no error")
	}

	// ...and succeeds once GitHub recovers.
	fake.mu.Lock()
	fake.failing["broken"] = false
	fake.mu.Unlock()
	hasReadme, err = s.GetHasReadme("broken")
	if err != nil || !hasReadme {
		t.Errorf("GetHasReadme() after recovery = %v, %v; want true, nil", hasReadme, err)
	}
}

// failingTransport fails every request below the HTTP level, as a refused
// connection or a DNS failure would, so go-github hands back a nil response.
type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("connection refused")
}

// A transport error reaches the loader with a nil response. It must come back
// as an error, not a nil dereference: inside a prefetch worker a panic would
// take the whole run down.
func TestContentDetails_TransportErrorIsAnError(t *testing.T) {
	client, err := github.NewClient(
		github.WithHTTPClient(&http.Client{Transport: failingTransport{}}),
		github.WithDisableRateLimitCheck(),
	)
	if err != nil {
		t.Fatal(err)
	}
	s := &Service{
		config:                   Config{GithubOrganization: "giantswarm"},
		ctx:                      context.Background(),
		githubClient:             client,
		githubRepoDetails:        make(map[string]GithubRepoDetails),
		githubRepoContentDetails: make(map[string]GithubRepoContentDetails),
	}

	s.PrefetchContentDetails([]string{"unreachable"}, 2)

	if _, err := s.GetHasReadme("unreachable"); err == nil {
		t.Error("GetHasReadme() against an unreachable API returned no error")
	}
}

func TestPrefetchContentDetails_Deduplicates(t *testing.T) {
	fake := &fakeContentsServer{requests: make(map[string]int)}
	s := newFakeService(t, fake)

	s.PrefetchContentDetails([]string{"solo"}, 1)
	once := fake.requestsFor("solo")

	s = newFakeService(t, fake)
	fake.requests = make(map[string]int)
	s.PrefetchContentDetails([]string{"shared", "shared", "shared"}, 4)

	if got := fake.requestsFor("shared"); got != once {
		t.Errorf("a repo listed three times issued %d requests, want %d (one load)", got, once)
	}
}

func TestGetChartYAML(t *testing.T) {
	fake := &fakeContentsServer{requests: make(map[string]int)}
	s := newFakeService(t, fake)

	s.PrefetchContentDetails([]string{"charted"}, 1)
	before := fake.requestsFor("charted")

	content, err := s.GetChartYAML("charted", "app")
	if err != nil || content != testChartYAML {
		t.Errorf("GetChartYAML() = %q, %v; want the chart's Chart.yaml", content, err)
	}
	if after := fake.requestsFor("charted"); after != before {
		t.Errorf("GetChartYAML() on a prefetched chart issued %d requests", after-before)
	}

	hasSchema, err := s.GetHasValuesSchema("charted")
	if err != nil || !hasSchema["app"] {
		t.Errorf("GetHasValuesSchema() = %v, %v; want app: true", hasSchema, err)
	}
}

// A rate limit pauses the pool until GitHub's reset instead of letting the
// remaining repos fail one after another against go-github's short circuit.
func TestPrefetchContentDetails_PausesOnRateLimit(t *testing.T) {
	fake := &fakeContentsServer{
		requests:      make(map[string]int),
		rateLimitOnce: map[string]bool{"repo-00": true},
	}
	s := newFakeService(t, fake)

	names := make([]string, 0, 20)
	for i := range 20 {
		names = append(names, fmt.Sprintf("repo-%02d", i))
	}

	const workers = 2
	start := time.Now()
	s.PrefetchContentDetails(names, workers)
	if elapsed := time.Since(start); elapsed < 900*time.Millisecond {
		t.Errorf("prefetch took %v, want it to wait out the 1s Retry-After", elapsed)
	}

	// Only repos already in flight when the limit hit can have failed.
	uncached := 0
	for _, name := range names {
		s.contentDetailsMu.RLock()
		_, ok := s.githubRepoContentDetails[name]
		s.contentDetailsMu.RUnlock()
		if !ok {
			uncached++
		}
	}
	if uncached > workers {
		t.Errorf("%d repos left uncached, want at most %d (those in flight when the limit hit)", uncached, workers)
	}
}
