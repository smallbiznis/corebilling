package idempotency

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultCacheTTL = 30 * time.Second

// Cache represents the Redis-based deduplication cache.
type Cache interface {
	SetHash(ctx context.Context, tenantID, key, hash string, ttl time.Duration) error
	GetHash(ctx context.Context, tenantID, key string) (string, error)
	Delete(ctx context.Context, tenantID, key string) error
}

// RedisCache implements Cache using Redis as backend.
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisCache creates a Redis-backed Cache with the provided TTL. If ttl is zero, a default is used.
func NewRedisCache(client *redis.Client, ttl time.Duration) *RedisCache {
	if ttl <= 0 {
		ttl = defaultCacheTTL
	}
	return &RedisCache{client: client, ttl: ttl}
}

func (c *RedisCache) SetHash(ctx context.Context, tenantID, key, hash string, ttl time.Duration) error {
	if c == nil || c.client == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = c.ttl
	}
	redisKey := cacheKey(tenantID, key)
	return c.client.Set(ctx, redisKey, hash, ttl).Err()
}

func (c *RedisCache) GetHash(ctx context.Context, tenantID, key string) (string, error) {
	if c == nil || c.client == nil {
		return "", nil
	}
	redisKey := cacheKey(tenantID, key)
	val, err := c.client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *RedisCache) Delete(ctx context.Context, tenantID, key string) error {
	if c == nil || c.client == nil {
		return nil
	}
	redisKey := cacheKey(tenantID, key)
	return c.client.Del(ctx, redisKey).Err()
}

func cacheKey(tenantID, key string) string {
	return fmt.Sprintf("idemp:%s:%s", tenantID, key)
}
