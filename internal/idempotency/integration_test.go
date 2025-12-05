package idempotency

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type integrationEnv struct {
	pool        *pgxpool.Pool
	cache       *RedisCache
	svc         *Service
	pg          testcontainers.Container
	redis       testcontainers.Container
	tenantID    string
	cacheTTL    time.Duration
	redisClient *redis.Client
}

func setupIntegration(t *testing.T) *integrationEnv {
	t.Helper()
	ctx := context.Background()

	requireDocker(t)

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15-alpine",
			Env:          map[string]string{"POSTGRES_PASSWORD": "password", "POSTGRES_USER": "user", "POSTGRES_DB": "testdb"},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)

	host, err := pgContainer.Host(ctx)
	require.NoError(t, err)
	port, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, fmt.Sprintf("postgres://user:password@%s:%s/testdb?sslmode=disable", host, port.Port()))
	require.NoError(t, err)
	applyMigrations(t, ctx, pool)

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)

	redisHost, err := redisContainer.Host(ctx)
	require.NoError(t, err)
	redisPort, err := redisContainer.MappedPort(ctx, "6379")
	require.NoError(t, err)

	redisClient := redis.NewClient(&redis.Options{Addr: redisHost + ":" + redisPort.Port()})
	cacheTTL := 2 * time.Second
	cache := NewRedisCache(redisClient, cacheTTL)

	repo := NewSQLRepository(pool)
	svc := NewService(repo, cache)

	return &integrationEnv{
		pool:        pool,
		cache:       cache,
		svc:         svc,
		pg:          pgContainer,
		redis:       redisContainer,
		tenantID:    "tenant",
		cacheTTL:    cacheTTL,
		redisClient: redisClient,
	}
}

func (e *integrationEnv) cleanup() {
	_ = e.redisClient.Close()
	e.pool.Close()
	ctx := context.Background()
	_ = e.pg.Terminate(ctx)
	_ = e.redis.Terminate(ctx)
}

func TestIntegrationFullFlow(t *testing.T) {
	env := setupIntegration(t)
	defer env.cleanup()

	ctx := context.Background()
	key := "order-1"
	body := []byte(`{"idempotency_key":"order-1","amount":100}`)

	// Case A: first request
	record, existing, err := env.svc.Begin(ctx, env.tenantID, key, body)
	require.NoError(t, err)
	require.False(t, existing)
	require.Equal(t, StatusProcessing, record.Status)

	response := map[string]interface{}{"status": "ok"}
	require.NoError(t, env.svc.Complete(ctx, env.tenantID, key, response))

	stored, err := env.svc.repo.Get(ctx, env.tenantID, key)
	require.NoError(t, err)
	require.Equal(t, StatusCompleted, stored.Status)

	// Case B: redis fast dedupe (processing state still cached)
	pendingKey := "pending-request"
	pendingBody := []byte(`{"idempotency_key":"pending-request"}`)
	record, existing, err = env.svc.Begin(ctx, env.tenantID, pendingKey, pendingBody)
	require.NoError(t, err)
	require.False(t, existing)
	require.Equal(t, StatusProcessing, record.Status)

	record, existing, err = env.svc.Begin(ctx, env.tenantID, pendingKey, pendingBody)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, StatusProcessing, record.Status)

	// Case C: redis TTL expired but PG completed
	time.Sleep(env.cacheTTL + 500*time.Millisecond)
	completedRecord, existing, err := env.svc.Begin(ctx, env.tenantID, key, body)
	require.NoError(t, err)
	require.True(t, existing)
	require.Equal(t, StatusCompleted, completedRecord.Status)
	require.NotEmpty(t, completedRecord.Response)

	// Case D: concurrent requests
	raceKey := "race-key"
	raceBody := []byte(`{"idempotency_key":"race-key"}`)
	var wg sync.WaitGroup
	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := env.svc.Begin(ctx, env.tenantID, raceKey, raceBody)
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	var count int
	err = env.pool.QueryRow(ctx, "SELECT COUNT(*) FROM idempotency_records WHERE key=$1", raceKey).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	// Case E: corrupted data (missing response)
	corruptKey := "corrupt"
	corruptHash := ComputeHash([]byte(`{"idempotency_key":"corrupt"}`))
	_, err = env.pool.Exec(ctx, "INSERT INTO idempotency_records (tenant_id, key, request_hash, status) VALUES ($1, $2, $3, 'COMPLETED') ON CONFLICT DO NOTHING", env.tenantID, corruptKey, corruptHash)
	require.NoError(t, err)

	corruptedRecord, existing, err := env.svc.Begin(ctx, env.tenantID, corruptKey, []byte(`{"idempotency_key":"corrupt"}`))
	require.NoError(t, err)
	require.True(t, existing)
	require.NotNil(t, corruptedRecord)
}
