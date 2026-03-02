package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter implements a token-bucket rate limiter using only the stdlib.
// Tokens are replenished at a fixed rate up to a maximum burst size.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	maxBurst float64
	rate     float64 // tokens per second
	lastTime time.Time
}

// NewLimiter creates a rate limiter that allows rps requests per second
// with an initial burst capacity.
func NewLimiter(rps float64, burst int) *Limiter {
	return &Limiter{
		tokens:   float64(burst),
		maxBurst: float64(burst),
		rate:     rps,
		lastTime: time.Now(),
	}
}

// Wait blocks until a token is available or the context is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		l.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(l.lastTime).Seconds()
		l.tokens += elapsed * l.rate
		if l.tokens > l.maxBurst {
			l.tokens = l.maxBurst
		}
		l.lastTime = now

		if l.tokens >= 1.0 {
			l.tokens -= 1.0
			l.mu.Unlock()
			return nil
		}

		// Calculate wait time for next token
		deficit := 1.0 - l.tokens
		wait := time.Duration(deficit / l.rate * float64(time.Second))
		l.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
}
