package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type RedisSearchCache struct {
	client *redis.Client
	prefix string // e.g "search_cache:"
}

func NewRedisSearchCache(redisClient *redis.Client, prefix string) *RedisSearchCache {
	return &RedisSearchCache{
		client: redisClient,
		prefix: prefix,
	}
}

func (r *RedisSearchCache) Get(ctx context.Context, cacheKey string) ([]byte, error) {
	if cacheKey == "" {
		return nil, fmt.Errorf("Invalid search key")
	}
	key := r.generateRedisKey(cacheKey)
	value, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("key: not found")
	}
	if err != nil {
		return nil, fmt.Errorf("Key: %s not found", key)
	}

	return []byte(value), nil
}

func (r *RedisSearchCache) Set(ctx context.Context, cacheKey string, data []byte, ttl time.Duration) error {
	if cacheKey == "" {
		return fmt.Errorf("Invalid search key")
	}
	key := r.generateRedisKey(cacheKey)

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("Unable to cache query. Error: %w", err)
	}

	return nil
}

func (r *RedisSearchCache) Clear(ctx context.Context) error {
	keys, err := r.client.Keys(ctx, fmt.Sprintf("%s:*", r.prefix)).Result()

	if err != nil {
		return fmt.Errorf("Unable to fetch redis keys: %w", err)
	}

	if len(keys) == 0 {
		log.Info().Msg("There are no cached queries to be deleted")
		return nil
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("Unable to clear cache. Error: %w", err)
	}

	log.Info().Msg("Cache cleared successfully!")
	return nil

}

func (r *RedisSearchCache) generateRedisKey(key string) string {
	return fmt.Sprintf("%s:%s", r.prefix, key)
}
