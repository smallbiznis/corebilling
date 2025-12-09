package interceptors

import (
	"context"
	"time"

	"github.com/smallbiznis/corebilling/internal/headers"
	"github.com/smallbiznis/corebilling/internal/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// MetricsUnaryInterceptor captures gRPC request metrics.
func MetricsUnaryInterceptor(metrics *telemetry.Metrics) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		status := "success"
		if err != nil {
			status = "error"
		}
		if metrics != nil {
			metrics.ObserveAPIRequest(info.FullMethod, status, extractTenant(ctx), time.Since(start))
		}
		return resp, err
	}
}

// MetricsStreamInterceptor captures gRPC stream metrics.
func MetricsStreamInterceptor(metrics *telemetry.Metrics) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, ss)
		status := "success"
		if err != nil {
			status = "error"
		}
		if metrics != nil {
			metrics.ObserveAPIRequest(info.FullMethod, status, extractTenant(ss.Context()), time.Since(start))
		}
		return err
	}
}

func extractTenant(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(headers.MetadataTenantID); len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}
