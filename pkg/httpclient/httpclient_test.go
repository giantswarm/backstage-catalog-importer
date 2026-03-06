package httpclient

import (
	"math"
	"net/http"
	"net/http/httptest"
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
	client := NewGitHubClient("test-token")
	if client == nil {
		t.Fatal("NewGitHubClient returned nil")
	}
}

func TestNewGitHubClient_NoToken(t *testing.T) {
	client := NewGitHubClient("")
	if client == nil {
		t.Fatal("NewGitHubClient with empty token returned nil")
	}
}
