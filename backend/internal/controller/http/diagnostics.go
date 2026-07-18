package router

import (
	"database/sql"
	"log/slog"
	"time"

	"cspirt/pkg/profiler"

	"github.com/gin-gonic/gin"
)

// DiagnosticsMiddleware measures every request and logs a timing breakdown plus
// database connection-pool pressure. It is the profiling layer described in the
// audit: total latency, per-layer spans, and — crucially — how long the request
// spent waiting for a pooled DB connection (db.WaitCount/WaitDuration deltas).
//
// A non-zero wait delta is the smoking gun for pool starvation: it means the
// handler blocked because all MaxOpenConns were busy. Enable with PROFILE=1.
func DiagnosticsMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		coll := profiler.New()
		c.Request = c.Request.WithContext(profiler.Inject(c.Request.Context(), coll))

		var (
			waitCountBefore int64
			waitDurBefore   time.Duration
		)
		if db != nil {
			s := db.Stats()
			waitCountBefore = s.WaitCount
			waitDurBefore = s.WaitDuration
		}

		start := time.Now()
		c.Next()
		total := time.Since(start)

		attrs := []any{
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"total", total.Round(time.Millisecond).String(),
		}
		if breakdown := coll.String(); breakdown != "" {
			attrs = append(attrs, "spans", breakdown)
		}
		if db != nil {
			s := db.Stats()
			if waited := s.WaitCount - waitCountBefore; waited > 0 {
				attrs = append(attrs,
					"db_wait_count", waited,
					"db_wait_time", (s.WaitDuration - waitDurBefore).Round(time.Millisecond).String(),
				)
			}
			attrs = append(attrs, "db_in_use", s.InUse, "db_idle", s.Idle)
		}

		slog.Info("request", attrs...)
	}
}
