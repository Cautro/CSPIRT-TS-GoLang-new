package repo

import (
	"context"
	"time"
)

// CacheRepository is the port for a key-value cache with TTL semantics.
// Any implementation (Redis, in-memory, ...) must satisfy this interface;
// usecases depend only on this interface, never on a concrete adapter.
type CacheRepository interface {
	// Set marshals value to JSON and stores it under key with the given ttl.
	// ttl <= 0 means "no expiration".
	Set(ctx context.Context, key string, value any, ttl time.Duration) error

	// Get unmarshals the JSON stored under key into dest (which must be a pointer).
	// Returns cache.ErrCacheMiss if the key does not exist.
	Get(ctx context.Context, key string, dest any) error

	// Delete removes one or more keys. Missing keys are ignored.
	Delete(ctx context.Context, keys ...string) error

	// Exists reports whether key is currently present.
	Exists(ctx context.Context, key string) (bool, error)

	// Increment atomically increments the integer counter stored at key and
	// returns the new value. If this is the first increment (the key was just
	// created), ttl is applied to the key so the counter resets automatically.
	Increment(ctx context.Context, key string, ttl time.Duration) (int64, error)

	// SetNX marshals value to JSON and stores it under key only if the key does
	// not already exist. Returns true if the key was created by this call.
	SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error)
}
