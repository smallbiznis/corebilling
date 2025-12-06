package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/smallbiznis/corebilling/internal/server/http/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module starts an HTTP server offering a readiness endpoint.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(middleware.LoggingMiddleware, fx.ResultTags(`group:"http_middleware"`)),
	),
	fx.Provide(
		fx.Annotate(middleware.RecoveryMiddleware, fx.ResultTags(`group:"http_middleware"`)),
	),
	fx.Provide(NewRegisterMux),
	fx.Invoke(Register),
)

type params struct {
	fx.In

	LC          fx.Lifecycle
	Mux         *runtime.ServeMux
	Logger      *zap.Logger
	Middlewares []middleware.Middleware `group:"http_middleware"`
}

func NewRegisterMux() *runtime.ServeMux {
	return runtime.NewServeMux()
}

// Register sets up lifecycle hooks for HTTP server.
func Register(p params) {
	mux := p.Mux

	mux.HandlePath("GET", "/ready", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	for i, mw := range p.Middlewares {
		p.Logger.Info("middleware loaded", zap.Int("index", i), zap.Reflect("mw", mw))
	}

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
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			p.Logger.Info("stopping http server")
			return srv.Shutdown(shutdownCtx)
		},
	})
}
