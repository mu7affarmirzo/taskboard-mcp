package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestLimiter_ImmediateBurst(t *testing.T) {
	l := NewLimiter(10, 3)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		start := time.Now()
		if err := l.Wait(ctx); err != nil {
			t.Fatalf("burst call %d failed: %v", i, err)
		}
		if time.Since(start) > 50*time.Millisecond {
			t.Fatalf("burst call %d took too long: %v", i, time.Since(start))
		}
	}
}

func TestLimiter_ThrottlesAfterBurst(t *testing.T) {
	l := NewLimiter(100, 1) // 100 rps, burst 1

	ctx := context.Background()
	// First call should be immediate (using burst token)
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Second call must wait for replenishment (~10ms at 100 rps)
	start := time.Now()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	elapsed := time.Since(start)
	if elapsed < 5*time.Millisecond {
		t.Fatalf("expected throttle delay, got %v", elapsed)
	}
}

func TestLimiter_ContextCancellation(t *testing.T) {
	l := NewLimiter(1, 0) // 0 burst, slow rate — will need to wait

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := l.Wait(ctx)
	if err == nil {
		t.Fatal("expected context error")
	}
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestLimiter_ContextTimeout(t *testing.T) {
	l := NewLimiter(0.1, 0) // Very slow rate, 0 burst

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := l.Wait(ctx)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestLimiter_ReplenishesOverTime(t *testing.T) {
	l := NewLimiter(100, 1)

	ctx := context.Background()
	// Exhaust burst
	if err := l.Wait(ctx); err != nil {
		t.Fatal(err)
	}

	// Wait for replenishment
	time.Sleep(50 * time.Millisecond)

	// Should have tokens available now
	start := time.Now()
	if err := l.Wait(ctx); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 10*time.Millisecond {
		t.Fatalf("expected immediate after replenishment, got %v", time.Since(start))
	}
}
