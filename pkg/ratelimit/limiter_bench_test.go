package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkLimiter_Wait_HighRate(b *testing.B) {
	l := NewLimiter(1_000_000, 1000) // very high rate to minimize blocking
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = l.Wait(ctx)
	}
}

func BenchmarkLimiter_Wait_Parallel(b *testing.B) {
	l := NewLimiter(1_000_000, 1000)
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = l.Wait(ctx)
		}
	})
}

func TestLimiter_ConcurrentAccess(t *testing.T) {
	l := NewLimiter(10_000, 100)
	ctx := context.Background()

	const goroutines = 50
	const callsPerGoroutine = 100

	var wg sync.WaitGroup
	var successCount atomic.Int64
	var errorCount atomic.Int64

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				if err := l.Wait(ctx); err != nil {
					errorCount.Add(1)
				} else {
					successCount.Add(1)
				}
			}
		}()
	}

	wg.Wait()

	total := successCount.Load() + errorCount.Load()
	if total != goroutines*callsPerGoroutine {
		t.Fatalf("expected %d total calls, got %d", goroutines*callsPerGoroutine, total)
	}
	if errorCount.Load() != 0 {
		t.Fatalf("expected 0 errors, got %d", errorCount.Load())
	}
}

func TestLimiter_BurstUnderConcurrency(t *testing.T) {
	// Low rate, high burst: all goroutines should get a token from the burst bucket
	const burst = 50
	l := NewLimiter(1, burst)
	ctx := context.Background()

	var wg sync.WaitGroup
	var fastCount atomic.Int64

	wg.Add(burst)
	start := time.Now()
	for i := 0; i < burst; i++ {
		go func() {
			defer wg.Done()
			if err := l.Wait(ctx); err != nil {
				return
			}
			if time.Since(start) < 100*time.Millisecond {
				fastCount.Add(1)
			}
		}()
	}

	wg.Wait()

	// All burst tokens should have been consumed quickly
	if fastCount.Load() < int64(burst/2) {
		t.Fatalf("expected at least %d fast calls from burst, got %d", burst/2, fastCount.Load())
	}
}

func TestLimiter_FairnessUnderContention(t *testing.T) {
	// Verify all goroutines eventually get tokens under contention
	l := NewLimiter(1000, 10)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const goroutines = 20
	counts := make([]atomic.Int64, goroutines)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				if err := l.Wait(ctx); err != nil {
					return
				}
				counts[idx].Add(1)
			}
		}(i)
	}

	wg.Wait()

	// Every goroutine should have completed all 50 calls
	for i := 0; i < goroutines; i++ {
		if counts[i].Load() < 50 {
			t.Errorf("goroutine %d only got %d tokens (expected 50)", i, counts[i].Load())
		}
	}
}
