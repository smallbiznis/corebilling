package middleware

import (
	"net/http"
	"time"

	"github.com/smallbiznis/corebilling/internal/headers"
	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
	"github.com/smallbiznis/corebilling/internal/telemetry/correlation"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var tracer = otel.Tracer("server.http")

// LoggingMiddleware ensures correlation-aware logging for HTTP requests.
func LoggingMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			corrHeader := r.Header.Get(headers.MetadataCorrelation)
			if corrHeader != "" {
				ctx = correlation.ContextWithCorrelationID(ctx, corrHeader)
			}
			ctx, cid := correlation.EnsureCorrelationID(ctx)
			ctx, span := tracer.Start(ctx, "http.request", trace.WithSpanKind(trace.SpanKindServer))
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.target", r.URL.Path),
				attribute.String("correlation_id", cid),
			)
			defer span.End()

			recorder := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
			r = r.WithContext(ctx)

			log := ctxlogger.FromContext(ctx)
			start := time.Now()
			log.Info("http.request.start", zap.String("method", r.Method), zap.String("path", r.URL.Path))

			next.ServeHTTP(recorder, r)

			duration := time.Since(start)
			log.Info("http.request.finish",
				zap.Int("status", recorder.status),
				zap.Int("bytes", recorder.size),
				zap.Duration("duration", duration),
			)
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}
