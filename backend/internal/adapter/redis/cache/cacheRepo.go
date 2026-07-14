// Package cache implements the internal/domain/cache/repo.CacheRepository port
// on top of Redis. Usecases should depend on the repo.CacheRepository
// interface, not on this package, so the backing store can be swapped without
// touching business logic.
package cache

import (
	"context"
	"encoding/json"
	"errors"

	"time"

	domainCache "cspirt/internal/domain/cache"
	"cspirt/internal/domain/cache/repo"

	goredis "github.com/redis/go-redis/v9"
)

type redisCacheRepository struct {
	client *goredis.Client
}

// New builds a repo.CacheRepository backed by an already-connected Redis
// client (see internal/adapter/redis.Client).
func New(client *goredis.Client) repo.CacheRepository {
	return &redisCacheRepository{client: client}
}

func (r *redisCacheRepository) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisCacheRepository) Get(ctx context.Context, key string, dest any) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domainCache.ErrCacheMiss
		}
		return err
	}
	return json.Unmarshal(data, dest)
}

func (r *redisCacheRepository) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *redisCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *redisCacheRepository) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	if count == 1 && ttl > 0 {
		if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
			return count, err
		}
	}

	return count, nil
}

func (r *redisCacheRepository) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, err
	}
	return r.client.SetNX(ctx, key, data, ttl).Result()
}
