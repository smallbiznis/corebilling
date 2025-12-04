package grpcserver

import (
	"context"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/smallbiznis/corebilling/internal/config"
)

// Module starts the gRPC server.
var Module = fx.Invoke(Register)

// Register wires lifecycle hooks for gRPC server.
func Register(lc fx.Lifecycle, cfg config.Config, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", ":50051")
			if err != nil {
				return err
			}

			s := grpc.NewServer(grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
			healthServer := health.NewServer()
			healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
			grpc_health_v1.RegisterHealthServer(s, healthServer)

			go func() {
				logger.Info("gRPC server listening", zap.String("addr", lis.Addr().String()))
				if err := s.Serve(lis); err != nil {
					logger.Error("gRPC server exited", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// graceful stop not necessary in sample
			return nil
		},
	})
}
