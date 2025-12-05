package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/config"
)

// Module wires the database connections via Fx.
var Module = fx.Provide(NewPool, NewSQLDB)

// NewPool creates a pgx connection pool.
func NewPool(lc fx.Lifecycle, cfg config.Config, logger *zap.Logger) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing database pool")
			pool.Close()
			return nil
		},
	})

	logger.Info("database connected")
	return pool, nil
}

// NewSQLDB exposes a database/sql handle for libraries that require it.
func NewSQLDB(lc fx.Lifecycle, cfg config.Config, logger *zap.Logger) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing sql.DB")
			return db.Close()
		},
	})
	logger.Info("sql.DB ready")
	return db, nil
}
