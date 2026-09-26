package repositories

import (
	"context"
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
// README.md and nothing else, except that repos named in failing return 500
// for their CircleCI config until healed.
type fakeContentsServer struct {
	mu       sync.Mutex
	requests map[string]int
	failing  map[string]bool

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
	f.mu.Unlock()

	switch {
	case failing && path == ".circleci/config.yml":
		http.Error(w, `{"message":"boom"}`, http.StatusInternalServerError)
	case path == "README.md":
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"type":"file","name":"README.md","path":"README.md","content":"","encoding":"base64"}`)
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"message":"Not Found"}`)
	}
}

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
	client, err := github.NewClient(
		github.WithHTTPClient(srv.Client()),
		github.WithURLs(&baseURL, &baseURL),
		github.WithDisableRateLimitCheck(),
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
