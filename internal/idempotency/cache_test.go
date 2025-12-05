package idempotency

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupRedis(t *testing.T) (*RedisCache, func()) {
	t.Helper()

	requireDocker(t)

	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{Addr: host + ":" + port.Port()})
	cache := NewRedisCache(client, 2*time.Second)

	cleanup := func() {
		_ = client.Close()
		_ = container.Terminate(ctx)
	}
	return cache, cleanup
}

func TestRedisCacheSetAndGet(t *testing.T) {
	cache, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	err := cache.SetHash(ctx, "tenant", "key", "hash", time.Second)
	require.NoError(t, err)

	val, err := cache.GetHash(ctx, "tenant", "key")
	require.NoError(t, err)
	require.Equal(t, "hash", val)
}

func TestRedisCacheOverwrite(t *testing.T) {
	cache, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, cache.SetHash(ctx, "tenant", "key", "first", time.Second))
	require.NoError(t, cache.SetHash(ctx, "tenant", "key", "second", time.Second))

	val, err := cache.GetHash(ctx, "tenant", "key")
	require.NoError(t, err)
	require.Equal(t, "second", val)
}

func TestRedisCacheTTLExpiration(t *testing.T) {
	cache, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, cache.SetHash(ctx, "tenant", "key", "temp", time.Second))

	time.Sleep(1200 * time.Millisecond)

	val, err := cache.GetHash(ctx, "tenant", "key")
	require.NoError(t, err)
	require.Equal(t, "", val)
}

func TestRedisCacheDelete(t *testing.T) {
	cache, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	require.NoError(t, cache.SetHash(ctx, "tenant", "key", "value", time.Second))
	require.NoError(t, cache.Delete(ctx, "tenant", "key"))

	val, err := cache.GetHash(ctx, "tenant", "key")
	require.NoError(t, err)
	require.Equal(t, "", val)
}

func TestRedisCacheMissingKey(t *testing.T) {
	cache, cleanup := setupRedis(t)
	defer cleanup()

	ctx := context.Background()
	val, err := cache.GetHash(ctx, "tenant", "missing")
	require.NoError(t, err)
	require.Equal(t, "", val)
}
