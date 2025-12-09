package httpserver

import (
	"context"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/smallbiznis/corebilling/internal/headers"
	"github.com/smallbiznis/corebilling/internal/server/http/middleware"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

// Module starts an HTTP server offering a readiness endpoint.
var Module = fx.Options(
	fx.Provide(
		fx.Annotate(middleware.LoggingMiddleware, fx.ResultTags(`group:"http_middleware"`)),
	),
	fx.Provide(
		fx.Annotate(middleware.RecoveryMiddleware, fx.ResultTags(`group:"http_middleware"`)),
	),
	fx.Provide(
		fx.Annotate(middleware.BillingMiddleware, fx.ResultTags(`group:"http_middleware"`)),
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
	marshaler := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			EmitUnpopulated: true,
			UseEnumNumbers:  false,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	return runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaler),
		runtime.WithIncomingHeaderMatcher(CustomHeaderMatcher),
		runtime.WithMetadata(APIKeyAnnotator),
		runtime.WithMetadata(TenantIDAnnotator),
		runtime.WithMetadata(UserIDAnnotator),
		runtime.WithMetadata(CorellationAnnotator),
		runtime.WithMetadata(IdempotencyKeyAnnotator),
		runtime.WithMetadata(EventIDAnnotator),
		runtime.WithMetadata(EventTypeAnnotator),
	)
}

func CustomHeaderMatcher(key string) (string, bool) {
	switch key {
	case headers.MetadataAPIKey:
		return headers.MetadataAPIKey, true
	case headers.MetadataTenantID:
		return headers.MetadataTenantID, true
	case headers.MetadataIdempotency:
		return headers.MetadataIdempotency, true
	case headers.MetadataCorrelation:
		return headers.MetadataCorrelation, true
	case headers.MetadataEventID:
		return headers.MetadataEventID, true
	case headers.MetadataEventType:
		return headers.MetadataEventType, true
	case headers.MetadataCausation:
		return headers.MetadataCausation, true
	case headers.MetadataSignature:
		return headers.MetadataSignature, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func APIKeyAnnotator(ctx context.Context, r *http.Request) metadata.MD {
	v := r.Header.Get(headers.HeaderAPIKey)
	if v == "" {
		return nil
	}
	return metadata.Pairs(headers.MetadataAPIKey, v)
}

func TenantIDAnnotator(ctx context.Context, r *http.Request) metadata.MD {
	v := r.Header.Get(headers.HeaderTenantID)
	if v == "" {
		return nil
	}
	return metadata.Pairs(headers.MetadataTenantID, v)
}

func UserIDAnnotator(ctx context.Context, r *http.Request) metadata.MD {
	v := r.Header.Get(headers.HeaderUserID)
	if v == "" {
		return nil
	}
	return metadata.Pairs(headers.MetadataUserID, v)
}

func IdempotencyKeyAnnotator(ctx context.Context, r *http.Request) metadata.MD {
	v := r.Header.Get(headers.HeaderIdempotency)
	if v == "" {
		return nil
	}
	return metadata.Pairs(headers.MetadataIdempotency, v)
}

func CorellationAnnotator(ctx context.Context, r *http.Request) metadata.MD {
	v := r.Header.Get(headers.HeaderCorrelation)
	if v == "" {
		return nil
	}
	return metadata.Pairs(headers.MetadataCorrelation, v)
}

func EventIDAnnotator(ctx context.Context, r *http.Request) metadata.MD {
	v := r.Header.Get(headers.HeaderEventID)
	if v == "" {
		return nil
	}
	return metadata.Pairs(headers.MetadataEventID, v)
}

func EventTypeAnnotator(ctx context.Context, r *http.Request) metadata.MD {
	v := r.Header.Get(headers.HeaderEventType)
	if v == "" {
		return nil
	}
	return metadata.Pairs(headers.MetadataEventType, v)
}

// Register sets up lifecycle hooks for HTTP server.
func Register(p params) {
	mux := p.Mux

	mux.HandlePath("GET", "/ready", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
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
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			p.Logger.Info("stopping http server")
			return srv.Shutdown(shutdownCtx)
		},
	})
}
