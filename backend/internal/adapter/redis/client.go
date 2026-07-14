// Package redis provides the low-level connection to Redis. It knows nothing
// about cache semantics or business logic — that lives in
// internal/adapter/redis/cache, which implements the internal/domain/cache/repo
// port on top of the *Client produced here.
package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Config holds Redis connection settings. Zero-value timeouts/pool size fall
// back to sane defaults in New.
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int

	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolSize     int
}

const (
	defaultDialTimeout  = 5 * time.Second
	defaultReadTimeout  = 3 * time.Second
	defaultWriteTimeout = 3 * time.Second
	defaultPoolSize     = 10
	defaultPingTimeout  = 5 * time.Second
)

// Client wraps the go-redis client so the rest of the codebase depends on
// this package's type instead of importing go-redis directly.
type Client struct {
	*goredis.Client
}

// New connects to Redis using cfg and verifies the connection with a PING.
func New(cfg Config) (*Client, error) {
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = defaultDialTimeout
	}
	if cfg.ReadTimeout <= 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}
	if cfg.WriteTimeout <= 0 {
		cfg.WriteTimeout = defaultWriteTimeout
	}
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = defaultPoolSize
	}

	rdb := goredis.NewClient(&goredis.Options{
		Addr:         cfg.Host + ":" + cfg.Port,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		PoolSize:     cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), defaultPingTimeout)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}

	return &Client{Client: rdb}, nil
}

// Close releases the underlying connection pool.
func (c *Client) Close() error {
	return c.Client.Close()
}
