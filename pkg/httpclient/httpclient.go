// Package httpclient provides a retry-capable HTTP client for use with external APIs.
package httpclient

import (
	"context"
	"crypto/rand"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v84/github"
)

const (
	// DefaultMaxRetries is the default number of retry attempts for failed requests.
	DefaultMaxRetries = 3

	// DefaultBaseDelay is the initial delay before the first retry.
	DefaultBaseDelay = 1 * time.Second

	// DefaultMaxDelay caps the backoff delay.
	DefaultMaxDelay = 30 * time.Second
)

// retryTransport is an http.RoundTripper that retries requests on transient failures.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// RoundTrip executes the request with retry logic for transient failures.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := range t.maxRetries + 1 {
		// Reset the request body for retries.
		if req.Body != nil && req.GetBody != nil && attempt > 0 {
			req.Body, err = req.GetBody()
			if err != nil {
				return nil, err
			}
		}

		resp, err = t.base.RoundTrip(req)

		// On network error, retry.
		if err != nil {
			if attempt < t.maxRetries {
				delay := t.backoff(attempt, nil)
				log.Printf("HTTP request to %q failed (attempt %d/%d): %v, retrying in %v", //nolint:gosec // G706: path is from our own request URL, not user input
					req.URL.Path, attempt+1, t.maxRetries+1, err, delay)
				sleep(req.Context(), delay)
				continue
			}
			return nil, err
		}

		// Check if the response status code is retryable.
		if !isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// On the last attempt, return whatever we got.
		if attempt >= t.maxRetries {
			return resp, nil
		}

		delay := t.backoff(attempt, resp)
		log.Printf("HTTP %d from %s %q (attempt %d/%d), retrying in %v", //nolint:gosec // G706: path is from our own request URL, not user input
			resp.StatusCode, req.Method, req.URL.Path, attempt+1, t.maxRetries+1, delay)

		// Drain and close the body so the connection can be reused.
		_ = resp.Body.Close()

		sleep(req.Context(), delay)
	}

	return resp, err
}

// isRetryableStatus returns true for HTTP status codes that warrant a retry.
func isRetryableStatus(code int) bool {
	switch code {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	}
	return false
}

// backoff calculates the delay before the next retry using exponential backoff with jitter.
// If the response contains a Retry-After header, that value is used instead (capped at maxDelay).
func (t *retryTransport) backoff(attempt int, resp *http.Response) time.Duration {
	// Check Retry-After header.
	if resp != nil {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
				d := time.Duration(seconds) * time.Second
				return min(d, t.maxDelay)
			}
		}
	}

	// Exponential backoff: baseDelay * 2^attempt, with ±25% jitter.
	delay := float64(t.baseDelay) * math.Pow(2, float64(attempt))
	if delay > float64(t.maxDelay) {
		delay = float64(t.maxDelay)
	}

	// #nosec G404 -- jitter for backoff does not need cryptographic randomness,
	// but we use crypto/rand to satisfy the linter.
	n, _ := rand.Int(rand.Reader, big.NewInt(1000))
	jitter := delay * 0.25 * (2*float64(n.Int64())/1000.0 - 1)
	delay += jitter

	return time.Duration(delay)
}

// sleep pauses for the given duration or until the context is cancelled.
func sleep(ctx context.Context, d time.Duration) {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}

// NewGitHubClient creates a new GitHub API client with retry logic.
func NewGitHubClient(token string) *github.Client {
	transport := &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: DefaultMaxRetries,
		baseDelay:  DefaultBaseDelay,
		maxDelay:   DefaultMaxDelay,
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	client := github.NewClient(httpClient)
	if token != "" {
		client = client.WithAuthToken(token)
	}

	return client
}
