package cache

import "errors"

// ErrCacheMiss is returned by CacheRepository.Get when the key does not exist.
var ErrCacheMiss = errors.New("cache: key not found")
