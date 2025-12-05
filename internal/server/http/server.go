package httpserver

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/smallbiznis/corebilling/internal/server/http/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module starts an HTTP server offering a readiness endpoint.
var Module = fx.Options(
	fx.Provide(fx.Annotate(middleware.LoggingMiddleware, fx.ResultTags(`group:"http_middleware"`))),
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
	// mux.HandlePath()
	// mux.HandleFunc("/ready", func(w http.ResponseWriter, _ *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// 	_, _ = w.Write([]byte("ok"))
	// })

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
