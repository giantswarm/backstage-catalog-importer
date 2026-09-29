// Package httpclient provides a retry-capable HTTP client for use with external APIs.
package httpclient

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v92/github"
)

const (
	// DefaultMaxRetries is the default number of retry attempts for failed requests.
	DefaultMaxRetries = 3

	// DefaultBaseDelay is the initial delay before the first retry.
	DefaultBaseDelay = 1 * time.Second

	// DefaultMaxDelay caps the backoff delay.
	DefaultMaxDelay = 30 * time.Second

	// DefaultMaxRateLimitDelay is the longest a rate-limited request waits
	// before its retry. A rate limit's wait is honoured in full, never cut
	// short: retrying before GitHub's window ends only meets the limit again,
	// and GitHub warns that clients which keep sending while limited can be
	// banned. A limit that asks for longer than this is returned at once
	// instead, since no retry within the budget could succeed.
	DefaultMaxRateLimitDelay = 2 * time.Minute

	// secondaryRateLimitDelay is the wait for a secondary rate limit that
	// names none. GitHub's guidance is to wait at least a minute.
	secondaryRateLimitDelay = time.Minute

	// rateLimitBodyPeek bounds how much of a 403/429 body is read to look for
	// GitHub's secondary rate limit message.
	rateLimitBodyPeek = 64 << 10
)

// retryTransport is an http.RoundTripper that retries requests on transient failures.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration

	// maxRateLimitDelay defaults to DefaultMaxRateLimitDelay when zero.
	maxRateLimitDelay time.Duration

	// sleep defaults to a context-aware real sleep; tests replace it.
	sleep func(ctx context.Context, d time.Duration)

	// now defaults to time.Now; tests replace it.
	now func() time.Time
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
				t.pause(req.Context(), delay)
				continue
			}
			return nil, err
		}

		// Check if the response warrants a retry, and after how long.
		delay, retry := t.retryDelay(req, attempt, resp)
		if !retry {
			return resp, nil
		}

		// On the last attempt, return whatever we got.
		if attempt >= t.maxRetries {
			return resp, nil
		}

		log.Printf("HTTP %d from %s %q (attempt %d/%d), retrying in %v", //nolint:gosec // G706: path is from our own request URL, not user input
			resp.StatusCode, req.Method, req.URL.Path, attempt+1, t.maxRetries+1, delay)

		// Drain and close the body so the connection can be reused.
		_ = resp.Body.Close()

		t.pause(req.Context(), delay)
	}

	return resp, err
}

// retryDelay reports whether resp warrants a retry and how long to wait
// first. A rate limit waits exactly as long as GitHub asks, or is not retried
// at all if that is beyond maxRateLimitDelay. Other transient failures back
// off exponentially.
func (t *retryTransport) retryDelay(req *http.Request, attempt int, resp *http.Response) (time.Duration, bool) {
	if wait, limited := t.rateLimitWait(resp); limited {
		maxWait := t.maxRateLimitDelay
		if maxWait == 0 {
			maxWait = DefaultMaxRateLimitDelay
		}
		if wait > maxWait {
			log.Printf("HTTP %d rate limit on %q asks to wait %v, more than %v, not retrying", //nolint:gosec // G706: path is from our own request URL, not user input
				resp.StatusCode, req.URL.Path, wait.Round(time.Second), maxWait)

			return 0, false
		}

		return wait, true
	}

	if isRetryableStatus(resp.StatusCode) {
		return t.backoff(attempt, resp), true
	}

	return 0, false
}

// rateLimitWait reports whether resp is a GitHub rate limit and, if so, how
// long GitHub asks to wait. GitHub signals it in three ways, checked in this
// order: a Retry-After header (seconds or an HTTP date); x-ratelimit-remaining
// at 0, with the window's end in x-ratelimit-reset; or, for some secondary
// limits, only the message in the body, which means waiting at least a
// minute. A 403 with none of these is a permission error, not a limit.
func (t *retryTransport) rateLimitWait(resp *http.Response) (time.Duration, bool) {
	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusTooManyRequests {
		return 0, false
	}

	now := time.Now
	if t.now != nil {
		now = t.now
	}

	if retryAfter := strings.TrimSpace(resp.Header.Get("Retry-After")); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			return time.Duration(max(seconds, 0)) * time.Second, true
		}
		if at, err := http.ParseTime(retryAfter); err == nil {
			return max(at.Sub(now()), 0), true
		}
	}

	if resp.Header.Get("X-RateLimit-Remaining") == "0" {
		if reset, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
			// One second of margin: the reset is given in whole seconds.
			return max(time.Unix(reset, 0).Sub(now())+time.Second, 0), true
		}
	}

	if bodyMentionsSecondaryRateLimit(resp) {
		return secondaryRateLimitDelay, true
	}

	return 0, false
}

// bodyMentionsSecondaryRateLimit peeks at the start of resp's body for
// GitHub's secondary rate limit message, leaving the body readable in full
// for whoever consumes the response.
func bodyMentionsSecondaryRateLimit(resp *http.Response) bool {
	if resp.Body == nil {
		return false
	}

	peek, err := io.ReadAll(io.LimitReader(resp.Body, rateLimitBodyPeek))
	resp.Body = struct {
		io.Reader
		io.Closer
	}{io.MultiReader(bytes.NewReader(peek), resp.Body), resp.Body}
	if err != nil {
		return false
	}

	return strings.Contains(strings.ToLower(string(peek)), "secondary rate limit")
}

// pause sleeps for d or until ctx is done.
func (t *retryTransport) pause(ctx context.Context, d time.Duration) {
	if t.sleep != nil {
		t.sleep(ctx, d)

		return
	}
	sleep(ctx, d)
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
func NewGitHubClient(token string) (*github.Client, error) {
	transport := &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: DefaultMaxRetries,
		baseDelay:  DefaultBaseDelay,
		maxDelay:   DefaultMaxDelay,
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	opts := []github.ClientOptionsFunc{github.WithHTTPClient(httpClient)}
	if token != "" {
		opts = append(opts, github.WithAuthToken(token))
	}

	return github.NewClient(opts...)
}
