package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/config"
)

// Module wires the database connection via Fx.
var Module = fx.Provide(NewPool)

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
