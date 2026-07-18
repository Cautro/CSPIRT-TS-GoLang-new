package storage

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Pool sizing. Defaults are sane for a small-to-mid deployment and can be
// overridden via env. The important relationship: /api/me fans out to 4
// concurrent queries per request, so MaxOpenConns must comfortably exceed
// 4 × expected-concurrent-/me-requests or requests queue for a connection
// (visible as db_wait_count in the diagnostics middleware).
const (
	defaultMaxOpenConns    = 25
	defaultMaxIdleConns    = 25
	defaultConnMaxLifetime = 5 * time.Minute
	// Keep idle connections warm for a while so a burst after a quiet period
	// doesn't pay TCP+TLS+auth dial latency on every connection — a source of
	// sporadic slow requests.
	defaultConnMaxIdleTime = 5 * time.Minute
)

func openPostgres(_ string) (*sql.DB, error) {
	db, err := sql.Open("pgx", os.Getenv("DB_PATH"))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", defaultMaxOpenConns))
	db.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", defaultMaxIdleConns))
	db.SetConnMaxLifetime(defaultConnMaxLifetime)
	db.SetConnMaxIdleTime(defaultConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func envInt(key string, def int) int {
	if raw := os.Getenv(key); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			return v
		}
	}
	return def
}
