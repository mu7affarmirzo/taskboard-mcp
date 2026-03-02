package httputil

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newTestClient(t *testing.T, server *httptest.Server, cfg RetryConfig) *http.Client {
	t.Helper()
	return &http.Client{
		Transport: &retryTransport{base: server.Client().Transport, config: cfg},
	}
}

func TestRetryClient_SuccessOnFirstAttempt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server, RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
	})

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestRetryClient_RetriesOnServerError(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server, RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   10 * time.Millisecond,
	})

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 after retries, got %d", resp.StatusCode)
	}
	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", calls.Load())
	}
}

func TestRetryClient_ReturnsLastResponseAfterMaxRetries(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := newTestClient(t, server, RetryConfig{
		MaxRetries: 2,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   5 * time.Millisecond,
	})

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", resp.StatusCode)
	}
	// 3 attempts (initial + 2 retries) + 1 final re-execute = 4
	if calls.Load() != 4 {
		t.Fatalf("expected 4 calls (3 retries + 1 final), got %d", calls.Load())
	}
}

func TestRetryClient_NoRetryOn4xx(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := newTestClient(t, server, RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   5 * time.Millisecond,
	})

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 call (no retry for 400), got %d", calls.Load())
	}
}

func TestRetryClient_RetriesOn429(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newTestClient(t, server, RetryConfig{
		MaxRetries: 2,
		BaseDelay:  1 * time.Millisecond,
		MaxDelay:   5 * time.Millisecond,
	})

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", calls.Load())
	}
}

func TestNewRetryHTTPClient(t *testing.T) {
	client := NewRetryHTTPClient(RetryConfig{MaxRetries: 1, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond}, 10*time.Second)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.Timeout != 10*time.Second {
		t.Fatalf("expected 10s timeout, got %v", client.Timeout)
	}
}

func TestIsRetryable(t *testing.T) {
	retryable := []int{429, 500, 502, 503, 504}
	for _, code := range retryable {
		if !isRetryable(code) {
			t.Errorf("expected %d to be retryable", code)
		}
	}
	nonRetryable := []int{200, 201, 301, 400, 401, 403, 404, 409}
	for _, code := range nonRetryable {
		if isRetryable(code) {
			t.Errorf("expected %d to NOT be retryable", code)
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	d := parseRetryAfter("5")
	if d == nil || *d != 5*time.Second {
		t.Fatal("expected 5s")
	}
	if parseRetryAfter("") != nil {
		t.Fatal("expected nil for empty")
	}
	if parseRetryAfter("not-a-number") != nil {
		t.Fatal("expected nil for non-numeric")
	}
}
