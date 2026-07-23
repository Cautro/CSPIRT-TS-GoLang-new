package profiler

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestCollectorConcurrentRecord exercises the exact access pattern of the
// /api/me fan-out: multiple goroutines record spans against one Collector held
// in a context. Run with `go test -race` on a platform that supports it to
// prove there is no data race on the shared collector.
func TestCollectorConcurrentRecord(t *testing.T) {
	c := New()
	ctx := Inject(context.Background(), c)

	const goroutines = 8
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = Track(ctx, "repo.work", func() error {
				time.Sleep(time.Millisecond)
				return nil
			})
		}(i)
	}
	wg.Wait()

	if got := len(c.Spans()); got != goroutines {
		t.Fatalf("expected %d spans, got %d", goroutines, got)
	}
}

// TestNilCollectorIsSafe verifies instrumentation is a no-op (never panics) when
// no collector is present, so profiling can be left off in production.
func TestNilCollectorIsSafe(t *testing.T) {
	var ran bool
	err := Track(context.Background(), "x", func() error { ran = true; return nil })
	if err != nil || !ran {
		t.Fatalf("Track without collector should run fn and return nil; ran=%v err=%v", ran, err)
	}
	if FromContext(context.Background()) != nil {
		t.Fatal("expected nil collector from bare context")
	}
}
