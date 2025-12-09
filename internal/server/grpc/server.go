package grpcserver

import (
	"context"
	"net"

	"github.com/smallbiznis/corebilling/internal/server/grpc/interceptors"
	"github.com/smallbiznis/corebilling/internal/telemetry"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Module starts the gRPC server and exposes shared server/health instances.
var Module = fx.Options(
	fx.Provide(NewServer),
	fx.Provide(NewHealthServer),
	fx.Invoke(RegisterHooks),
)

// NewServer constructs the shared gRPC server instance.
func NewServer(metrics *telemetry.Metrics) *grpc.Server {
	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.MetricsUnaryInterceptor(metrics),
			interceptors.LoggingUnaryInterceptor,
			interceptors.BillingUnaryInterceptor,
		),
		grpc.ChainStreamInterceptor(
			interceptors.MetricsStreamInterceptor(metrics),
			interceptors.LoggingStreamInterceptor,
			interceptors.BillingStreamInterceptor,
		),
		// grpc.StatsHandler(),
	)
}

// NewHealthServer builds a health server set to SERVING.
func NewHealthServer() *health.Server {
	srv := health.NewServer()
	srv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	return srv
}

// RegisterHooks wires lifecycle hooks for gRPC server.
func RegisterHooks(lc fx.Lifecycle, logger *zap.Logger, server *grpc.Server, healthServer *health.Server) {
	grpc_health_v1.RegisterHealthServer(server, healthServer)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", ":50052")
			if err != nil {
				return err
			}

			go func() {
				logger.Info("gRPC server listening", zap.String("addr", lis.Addr().String()))
				if err := server.Serve(lis); err != nil {
					logger.Error("gRPC server exited", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("stopping gRPC server")
			server.GracefulStop()
			return nil
		},
	})
}
