// Package profiler provides lightweight per-request timing. A Collector is
// stashed in the request context; any layer (handler, usecase, repository) can
// open a named span around a unit of work. The HTTP middleware prints the tree
// on request completion so you can see exactly where a request spends its time:
//
//	GET /api/me total=35ms
//	  usecase.GetFullUserInfo 31ms
//	    repo.notes 2ms
//	    repo.events 20ms
//	    repo.complaints 3ms
//	    cache.get 1ms
//
// It is intentionally allocation-cheap and concurrency-safe: the fan-out in
// GetFullUserInfoByLogin records spans from four goroutines at once.
package profiler

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

type ctxKey struct{}

// Span is a single measured section of work.
type Span struct {
	Name    string
	Elapsed time.Duration
}

// Collector accumulates spans for one request. The zero value is not usable;
// build one with New.
type Collector struct {
	mu    sync.Mutex
	spans []Span
}

// New returns an empty Collector.
func New() *Collector {
	return &Collector{spans: make([]Span, 0, 8)}
}

// Inject stores c in ctx so downstream layers can find it with FromContext.
func Inject(ctx context.Context, c *Collector) context.Context {
	return context.WithValue(ctx, ctxKey{}, c)
}

// FromContext returns the Collector carried by ctx, or nil if none. Callers
// must tolerate nil (profiling disabled) so instrumentation never changes
// behaviour.
func FromContext(ctx context.Context) *Collector {
	if ctx == nil {
		return nil
	}
	c, _ := ctx.Value(ctxKey{}).(*Collector)
	return c
}

// record appends a finished span. Safe for concurrent use.
func (c *Collector) record(name string, d time.Duration) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.spans = append(c.spans, Span{Name: name, Elapsed: d})
	c.mu.Unlock()
}

// Track measures fn and records it under name against the collector in ctx.
// If ctx carries no collector it just runs fn — zero overhead beyond the call.
func Track(ctx context.Context, name string, fn func() error) error {
	c := FromContext(ctx)
	if c == nil {
		return fn()
	}
	start := time.Now()
	err := fn()
	c.record(name, time.Since(start))
	return err
}

// Spans returns a copy of the recorded spans, slowest first.
func (c *Collector) Spans() []Span {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	out := make([]Span, len(c.spans))
	copy(out, c.spans)
	c.mu.Unlock()

	sort.Slice(out, func(i, j int) bool {
		return out[i].Elapsed > out[j].Elapsed
	})
	return out
}

// String renders the spans as a compact one-line breakdown, e.g.
// "repo.events=20ms repo.complaints=3ms repo.notes=2ms cache.get=1ms".
func (c *Collector) String() string {
	spans := c.Spans()
	if len(spans) == 0 {
		return ""
	}
	var b strings.Builder
	for i, s := range spans {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s.Name)
		b.WriteByte('=')
		b.WriteString(s.Elapsed.Round(time.Millisecond).String())
	}
	return b.String()
}
