package httpclient

import (
	"context"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryTransport_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond,
		maxDelay:   100 * time.Millisecond,
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRetryTransport_RetriesOn500(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond,
		maxDelay:   100 * time.Millisecond,
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if got := attempts.Load(); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestRetryTransport_RetriesOn504(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusGatewayTimeout)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond,
		maxDelay:   100 * time.Millisecond,
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if got := attempts.Load(); got != 2 {
		t.Errorf("expected 2 attempts, got %d", got)
	}
}

func TestRetryTransport_RetriesOn429WithRetryAfter(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport := &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond,
		maxDelay:   5 * time.Second,
	}
	client := &http.Client{Transport: transport}

	start := time.Now()
	resp, err := client.Get(server.URL)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if elapsed < 800*time.Millisecond {
		t.Errorf("expected at least 800ms delay for Retry-After, got %v", elapsed)
	}
}

func TestRetryTransport_ExhaustsRetries(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	transport := &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: 2,
		baseDelay:  10 * time.Millisecond,
		maxDelay:   100 * time.Millisecond,
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", resp.StatusCode)
	}
	// 1 initial + 2 retries = 3
	if got := attempts.Load(); got != 3 {
		t.Errorf("expected 3 attempts, got %d", got)
	}
}

func TestRetryTransport_NoRetryOn4xx(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	transport := &retryTransport{
		base:       http.DefaultTransport,
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond,
		maxDelay:   100 * time.Millisecond,
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got := attempts.Load(); got != 1 {
		t.Errorf("expected 1 attempt for 404, got %d", got)
	}
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{200, false},
		{201, false},
		{301, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}
	for _, tt := range tests {
		if got := isRetryableStatus(tt.code); got != tt.want {
			t.Errorf("isRetryableStatus(%d) = %v, want %v", tt.code, got, tt.want)
		}
	}
}

func TestBackoff_ExponentialWithJitter(t *testing.T) {
	transport := &retryTransport{
		baseDelay: 100 * time.Millisecond,
		maxDelay:  10 * time.Second,
	}

	for attempt := range 5 {
		delay := transport.backoff(attempt, nil)
		expectedBase := float64(100*time.Millisecond) * math.Pow(2, float64(attempt))
		if expectedBase > float64(10*time.Second) {
			expectedBase = float64(10 * time.Second)
		}
		// Jitter is ±25%, so delay should be within 75%-125% of base
		low := time.Duration(expectedBase * 0.75)
		high := time.Duration(expectedBase * 1.25)
		if delay < low || delay > high {
			t.Errorf("attempt %d: delay %v not in range [%v, %v]", attempt, delay, low, high)
		}
	}
}

func TestBackoff_RespectsRetryAfterHeader(t *testing.T) {
	transport := &retryTransport{
		baseDelay: 100 * time.Millisecond,
		maxDelay:  10 * time.Second,
	}

	resp := &http.Response{
		Header: http.Header{},
	}
	resp.Header.Set("Retry-After", "5")

	delay := transport.backoff(0, resp)
	if delay != 5*time.Second {
		t.Errorf("expected 5s from Retry-After, got %v", delay)
	}
}

func TestBackoff_CapsRetryAfterAtMaxDelay(t *testing.T) {
	transport := &retryTransport{
		baseDelay: 100 * time.Millisecond,
		maxDelay:  3 * time.Second,
	}

	resp := &http.Response{
		Header: http.Header{},
	}
	resp.Header.Set("Retry-After", "60")

	delay := transport.backoff(0, resp)
	if delay != 3*time.Second {
		t.Errorf("expected max delay 3s, got %v", delay)
	}
}

func TestNewGitHubClient(t *testing.T) {
	client, err := NewGitHubClient("test-token")
	if err != nil {
		t.Fatalf("NewGitHubClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewGitHubClient returned nil")
	}
}

func TestNewGitHubClient_NoToken(t *testing.T) {
	client, err := NewGitHubClient("")
	if err != nil {
		t.Fatalf("NewGitHubClient with empty token returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewGitHubClient with empty token returned nil")
	}
}

func TestRetryTransport_RateLimits(t *testing.T) {
	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		status       int
		headers      map[string]string
		body         string
		wantAttempts int32
		wantStatus   int
		wantDelay    time.Duration
	}{
		{
			name:         "secondary rate limit with Retry-After seconds",
			status:       http.StatusForbidden,
			headers:      map[string]string{"Retry-After": "1"},
			wantAttempts: 2,
			wantStatus:   http.StatusOK,
			wantDelay:    time.Second,
		},
		{
			name:         "Retry-After beyond maxDelay is honoured in full, not capped",
			status:       http.StatusForbidden,
			headers:      map[string]string{"Retry-After": "60"},
			wantAttempts: 2,
			wantStatus:   http.StatusOK,
			wantDelay:    60 * time.Second,
		},
		{
			name:         "Retry-After as an HTTP date",
			status:       http.StatusTooManyRequests,
			headers:      map[string]string{"Retry-After": now.Add(90 * time.Second).Format(http.TimeFormat)},
			wantAttempts: 2,
			wantStatus:   http.StatusOK,
			wantDelay:    90 * time.Second,
		},
		{
			name:   "x-ratelimit-remaining 0 waits until the reset",
			status: http.StatusForbidden,
			headers: map[string]string{
				"X-RateLimit-Remaining": "0",
				"X-RateLimit-Reset":     strconv.FormatInt(now.Add(20*time.Second).Unix(), 10),
			},
			wantAttempts: 2,
			wantStatus:   http.StatusOK,
			wantDelay:    21 * time.Second,
		},
		{
			name:   "a reset beyond the rate limit budget is returned at once",
			status: http.StatusForbidden,
			headers: map[string]string{
				"X-RateLimit-Remaining": "0",
				"X-RateLimit-Reset":     strconv.FormatInt(now.Add(time.Hour).Unix(), 10),
			},
			wantAttempts: 1,
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "secondary rate limit named only in the body waits a minute",
			status:       http.StatusForbidden,
			body:         `{"message":"You have exceeded a secondary rate limit. Please wait a few minutes before you try again."}`,
			wantAttempts: 2,
			wantStatus:   http.StatusOK,
			wantDelay:    time.Minute,
		},
		{
			name:         "permission error is not retried",
			status:       http.StatusForbidden,
			body:         `{"message":"Resource not accessible by integration"}`,
			wantAttempts: 1,
			wantStatus:   http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attempts atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if attempts.Add(1) == 1 {
					for k, v := range tt.headers {
						w.Header().Set(k, v)
					}
					w.WriteHeader(tt.status)
					_, _ = w.Write([]byte(tt.body))
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			var delays []time.Duration
			transport := &retryTransport{
				base:       http.DefaultTransport,
				maxRetries: 3,
				baseDelay:  10 * time.Millisecond,
				maxDelay:   5 * time.Second,
				sleep:      func(_ context.Context, d time.Duration) { delays = append(delays, d) },
				now:        func() time.Time { return now },
			}
			client := &http.Client{Transport: transport}

			resp, err := client.Get(server.URL)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.wantStatus)
			}
			if got := attempts.Load(); got != tt.wantAttempts {
				t.Errorf("attempts = %d, want %d", got, tt.wantAttempts)
			}
			if tt.wantAttempts > 1 && (len(delays) != 1 || delays[0] != tt.wantDelay) {
				t.Errorf("delays = %v, want [%v]", delays, tt.wantDelay)
			}
			if tt.wantAttempts == 1 && tt.body != "" && string(body) != tt.body {
				t.Errorf("peeking at the body consumed it: got %q, want %q", body, tt.body)
			}
		})
	}
}
