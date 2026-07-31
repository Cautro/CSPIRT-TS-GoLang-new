package router

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var RequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests",
	},
	[]string{"method", "path", "status"},
)

var RequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Request duration",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path"},
)

var ActiveRequests = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "http_active_requests",
        Help: "Current active requests",
    },
)

func RegisterMetrics() {
    prometheus.MustRegister(RequestsTotal)
    prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(ActiveRequests)
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ActiveRequests.Inc()
        defer ActiveRequests.Dec()

		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		status := strconv.Itoa(c.Writer.Status())

		RequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		RequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)
	}
}