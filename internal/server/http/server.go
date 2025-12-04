package httpserver

import (
	"context"
	"net/http"

	"github.com/smallbiznis/corebilling/internal/server/http/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module starts an HTTP server offering a readiness endpoint.
var Module = fx.Invoke(Register)

// Register sets up lifecycle hooks for HTTP server.
func Register(lc fx.Lifecycle, logger *zap.Logger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: middleware.LoggingMiddleware(mux),
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("http server listening", zap.String("addr", srv.Addr))
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Error("http server exited", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping http server")
			return srv.Shutdown(ctx)
		},
	})
}
