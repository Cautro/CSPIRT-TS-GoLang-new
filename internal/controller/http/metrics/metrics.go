package router

import (
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

var ErrorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_errors_total",
		Help: "Total HTTP errors (status >= 400)",
	},
	[]string{"method", "path", "status", "error_class"}, // error_class: "4xx" или "5xx"
)

var CacheOperationsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "cache_operations_total",
		Help: "Total number of cache operations and their outcome",
	},
	[]string{"operation", "status"},
)

var CacheOperationDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "cache_operation_duration_seconds",
		Help:    "Latency of cache operations in seconds",
		Buckets: []float64{0.0005, 0.001, 0.005, 0.01, 0.05, 0.1}, // шкала от 0.5мс до 100мс
	},
	[]string{"operation"},
)

var AuthAttempts = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "auth_attempts_total",
		Help: "Total authentication and token refresh attempts",
	},
	[]string{"type", "status"},
)

var GlobalEventsActionsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "global_events_actions_total",
		Help: "Total actions performed on global events",
	},
	[]string{"action"},
)

func RegisterMetrics() {
    prometheus.MustRegister(RequestsTotal)
    prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(ActiveRequests)
	prometheus.MustRegister(ErrorsTotal)
	prometheus.MustRegister(CacheOperationDuration)
	prometheus.MustRegister(CacheOperationsTotal)
	prometheus.MustRegister(AuthAttempts)
	prometheus.MustRegister(GlobalEventsActionsTotal)
}

