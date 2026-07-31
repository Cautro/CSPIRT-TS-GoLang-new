package router

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

func RegisterDBMetrics(db *sql.DB) {
	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_open_connections",
			Help: "Number of established connections to the database",
		},
		func() float64 { return float64(db.Stats().OpenConnections) },
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_in_use_connections",
			Help: "Number of connections currently in use",
		},
		func() float64 { return float64(db.Stats().InUse) },
	))

	prometheus.MustRegister(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_wait_duration_seconds_total",
			Help: "Total time spent waiting for a new connection",
		},
		func() float64 { return db.Stats().WaitDuration.Seconds() },
	))
}