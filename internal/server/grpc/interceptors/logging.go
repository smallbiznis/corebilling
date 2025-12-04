package interceptors

import (
	"context"

	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
	"github.com/smallbiznis/corebilling/internal/telemetry/correlation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var tracer = otel.Tracer("server.grpc")

// LoggingUnaryInterceptor adds correlation-aware logging around gRPC requests.
func LoggingUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	ctx, _ = correlation.EnsureCorrelationID(ctx)
	ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	log := ctxlogger.FromContext(ctx)
	log.Info("grpc.request.start", zap.Any("request", req))

	resp, err = handler(ctx, req)

	if err != nil {
		log.Error("grpc.request.error", zap.Error(err))
	} else {
		log.Info("grpc.request.finish")
	}

	return resp, err
}

// LoggingStreamInterceptor adds correlation-aware logging around gRPC streams.
func LoggingStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx, _ := correlation.EnsureCorrelationID(ss.Context())
	ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	log := ctxlogger.FromContext(ctx)
	log.Info("grpc.stream.start")

	wrapped := &serverStream{ServerStream: ss, ctx: ctx}
	err := handler(srv, wrapped)
	if err != nil {
		log.Error("grpc.stream.error", zap.Error(err))
	} else {
		log.Info("grpc.stream.finish")
	}
	return err
}

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *serverStream) Context() context.Context {
	return s.ctx
}
