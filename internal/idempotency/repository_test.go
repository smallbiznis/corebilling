package idempotency

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupPostgres(t *testing.T) (*pgxpool.Pool, testcontainers.Container, func()) {
	t.Helper()

	requireDocker(t)

	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:15-alpine",
			Env:          map[string]string{"POSTGRES_PASSWORD": "password", "POSTGRES_USER": "user", "POSTGRES_DB": "testdb"},
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort("5432/tcp"),
		},
		Started: true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://user:password@%s:%s/testdb?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)

	applyMigrations(t, ctx, pool)

	cleanup := func() {
		pool.Close()
		_ = container.Terminate(ctx)
	}
	return pool, container, cleanup
}

func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	root := projectRoot(t)
	migrationPath := filepath.Join(root, "db", "migrations", "idempotency", "up", "000001_idempotency.sql")

	migration, err := os.ReadFile(migrationPath)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, string(migration))
	require.NoError(t, err)
}

func projectRoot(t *testing.T) string {
	t.Helper()

	cwd, err := os.Getwd()
	require.NoError(t, err)

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		} else if !os.IsNotExist(err) {
			require.NoError(t, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	t.Fatalf("project root not found from %s", cwd)
	return ""
}

func TestSQLRepositoryInsertAndGet(t *testing.T) {
	pool, _, cleanup := setupPostgres(t)
	defer cleanup()

	repo := NewSQLRepository(pool)

	ctx := context.Background()
	err := repo.InsertProcessing(ctx, "tenant", "key1", "hash1")
	require.NoError(t, err)

	record, err := repo.Get(ctx, "tenant", "key1")
	require.NoError(t, err)
	require.NotNil(t, record)
	require.Equal(t, StatusProcessing, record.Status)
	require.Equal(t, "hash1", record.RequestHash)
}

func TestSQLRepositoryMarkCompleted(t *testing.T) {
	pool, _, cleanup := setupPostgres(t)
	defer cleanup()

	repo := NewSQLRepository(pool)

	ctx := context.Background()
	require.NoError(t, repo.InsertProcessing(ctx, "tenant", "key2", "hash2"))

	payload := []byte(`{"ok":true}`)
	require.NoError(t, repo.MarkCompleted(ctx, "tenant", "key2", payload))

	record, err := repo.Get(ctx, "tenant", "key2")
	require.NoError(t, err)
	require.Equal(t, StatusCompleted, record.Status)
	require.JSONEq(t, string(payload), string(record.Response))
}

func TestSQLRepositoryGetMissing(t *testing.T) {
	pool, _, cleanup := setupPostgres(t)
	defer cleanup()

	repo := NewSQLRepository(pool)

	ctx := context.Background()
	record, err := repo.Get(ctx, "tenant", "missing")
	require.Error(t, err)
	require.Nil(t, record)
}
