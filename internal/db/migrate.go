package db

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/config"
)

// MigrationsModule runs migrations during startup.
var MigrationsModule = fx.Invoke(RunMigrations)

// RunMigrations applies all enabled service migrations on startup.
func RunMigrations(lc fx.Lifecycle, cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return applyMigrations(ctx, cfg, logger, pool)
		},
	})
}

func applyMigrations(ctx context.Context, cfg config.Config, logger *zap.Logger, pool *pgxpool.Pool) error {
	root := cfg.MigrationsRoot
	for _, service := range cfg.EnabledMigrationServices {
		upDir := filepath.Join(root, service, "up")
		entries, err := os.ReadDir(upDir)
		if err != nil {
			if os.IsNotExist(err) {
				logger.Warn("migration directory missing", zap.String("service", service), zap.String("path", upDir))
				continue
			}
			return fmt.Errorf("read migrations: %w", err)
		}

		files := filterSQL(entries)
		sort.Strings(files)
		for _, file := range files {
			path := filepath.Join(upDir, file)
			if err := applyFile(ctx, pool, path); err != nil {
				logger.Error("migration failed", zap.String("service", service), zap.String("file", file), zap.Error(err))
				return err
			}
			logger.Info("migration applied", zap.String("service", service), zap.String("file", file))
		}
	}
	return nil
}

func filterSQL(entries []fs.DirEntry) []string {
	files := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	return files
}

func applyFile(ctx context.Context, pool *pgxpool.Pool, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read file %s: %w", path, err)
	}

	tctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	_, err = pool.Exec(tctx, string(content))
	if err != nil {
		return fmt.Errorf("exec migration %s: %w", path, err)
	}
	return nil
}
