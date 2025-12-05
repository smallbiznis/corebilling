package httpserver

import (
	"context"
	"net/http"

	"github.com/smallbiznis/corebilling/internal/server/http/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module starts an HTTP server offering a readiness endpoint.
var Module = fx.Options(
	fx.Provide(fx.Annotate(middleware.LoggingMiddleware, fx.ResultTags(`group:"http_middleware"`))),
	fx.Invoke(Register),
)

type params struct {
	fx.In

	LC          fx.Lifecycle
	Logger      *zap.Logger
	Middlewares []middleware.Middleware `group:"http_middleware"`
}

// Register sets up lifecycle hooks for HTTP server.
func Register(p params) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := middleware.Chain(mux, p.Middlewares...)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				p.Logger.Info("http server listening", zap.String("addr", srv.Addr))
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					p.Logger.Error("http server exited", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Logger.Info("stopping http server")
			return srv.Shutdown(ctx)
		},
	})
}
