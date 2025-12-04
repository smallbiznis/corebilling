package router

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
)

// Router wires subscriptions to handlers across providers.
type Router struct {
	bus      events.Bus
	logger   *zap.Logger
	handlers map[string]events.Handler
	group    string
	tracer   trace.Tracer
}

// NewRouter constructs a router with a consumer group identifier.
func NewRouter(bus events.Bus, logger *zap.Logger, group string) *Router {
	return &Router{bus: bus, logger: logger, handlers: make(map[string]events.Handler), group: group, tracer: otel.Tracer("events.router")}
}

// Register attaches a handler for a subject.
func (r *Router) Register(subject string, handler events.Handler) {
	r.handlers[subject] = r.wrapHandler(subject, handler)
}

// Start subscribes all registered handlers.
func (r *Router) Start(ctx context.Context) error {
	for subject, handler := range r.handlers {
		if err := r.bus.Subscribe(ctx, subject, r.group, handler); err != nil {
			return err
		}
	}
	r.logger.Info("event router started", zap.Int("handler_count", len(r.handlers)), zap.String("group", r.group))
	return nil
}

func (r *Router) wrapHandler(subject string, h events.Handler) events.Handler {
	return func(ctx context.Context, evt *events.Event) error {
		corr := ""
		if evt != nil && evt.Metadata != nil {
			if val, ok := evt.Metadata.Fields["correlation_id"]; ok {
				corr = val.GetStringValue()
			}
		}
		ctx = ctxlogger.ContextWithEventSubject(ctx, subject)
		ctx, span := r.tracer.Start(ctx, "event.handle", trace.WithAttributes(
			attribute.String("event.subject", subject),
			attribute.String("event.correlation_id", corr),
		))
		defer span.End()
		return h(ctx, evt)
	}
}
