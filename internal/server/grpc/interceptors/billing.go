package interceptors

import (
	"context"

	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
	"github.com/smallbiznis/corebilling/internal/telemetry/correlation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func BillingUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	// correlation ID
	ctx, _ = correlation.EnsureCorrelationID(ctx)

	// tracing
	ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	log := ctxlogger.FromContext(ctx)
	log.Info("grpc.request.start", zap.Any("request", req))

	// ⚠️ wajib panggil handler
	resp, err = handler(ctx, req)

	// log hasil
	if err != nil {
		log.Error("grpc.request.error", zap.Error(err))
		span.RecordError(err)
	} else {
		log.Info("grpc.request.end", zap.Any("response", resp))
	}

	return resp, err
}

func BillingStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
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
