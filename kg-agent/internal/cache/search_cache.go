package cache

import (
	"context"
	"time"
)

type SearchCache interface {
	// Get cached search result
	Get(ctx context.Context, cacheKey string) ([]byte, error)
	// Set search results in cache with ttl
	Set(ctx context.Context, cacheKey string, data []byte, ttl time.Duration) error
	// Clear all the cache
	Clear(ctx context.Context) error
}
