package httputil

import (
	"math"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

// NewRetryHTTPClient returns a standard *http.Client whose Transport
// automatically retries on transient errors (429, 500, 502, 503, 504)
// with exponential backoff and jitter.
func NewRetryHTTPClient(cfg RetryConfig, timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: &retryTransport{base: http.DefaultTransport, config: cfg},
	}
}

type retryTransport struct {
	base   http.RoundTripper
	config RetryConfig
}

func (rt *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= rt.config.MaxRetries; attempt++ {
		resp, err = rt.base.RoundTrip(req)
		if err != nil {
			if attempt == rt.config.MaxRetries {
				return nil, err
			}
			sleep(rt.config, attempt, nil)
			continue
		}

		if !isRetryable(resp.StatusCode) {
			return resp, nil
		}

		// Retryable status code — close body before retry
		retryAfter := resp.Header.Get("Retry-After")
		_ = resp.Body.Close()

		if attempt == rt.config.MaxRetries {
			// Final attempt — re-execute to return response to caller
			return rt.base.RoundTrip(req)
		}

		sleep(rt.config, attempt, parseRetryAfter(retryAfter))
	}

	return resp, err
}

func isRetryable(status int) bool {
	switch status {
	case http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}

func sleep(cfg RetryConfig, attempt int, retryAfter *time.Duration) {
	if retryAfter != nil {
		time.Sleep(*retryAfter)
		return
	}
	backoff := float64(cfg.BaseDelay) * math.Pow(2, float64(attempt))
	if backoff > float64(cfg.MaxDelay) {
		backoff = float64(cfg.MaxDelay)
	}
	// Add jitter: 50%-100% of calculated backoff
	jitter := backoff * (0.5 + rand.Float64()*0.5)
	time.Sleep(time.Duration(jitter))
}

func parseRetryAfter(value string) *time.Duration {
	if value == "" {
		return nil
	}
	seconds, err := strconv.Atoi(value)
	if err != nil {
		return nil
	}
	d := time.Duration(seconds) * time.Second
	return &d
}
