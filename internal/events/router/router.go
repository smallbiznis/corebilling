package router

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
	"github.com/smallbiznis/corebilling/internal/telemetry"
)

// Router wires subscriptions to handlers across providers.
type Router struct {
	bus      events.Bus
	logger   *zap.Logger
	handlers map[string]events.Handler
	group    string
	tracer   trace.Tracer
	metrics  *telemetry.Metrics
	sem      chan struct{}
	pre      []func(context.Context, *events.Event) error
}

// NewRouter constructs a router with a consumer group identifier.
func NewRouter(bus events.Bus, logger *zap.Logger, group string, metrics *telemetry.Metrics) *Router {
	concurrency := handlerConcurrency()
	return &Router{
		bus:      bus,
		logger:   logger,
		handlers: make(map[string]events.Handler),
		group:    group,
		tracer:   otel.Tracer("events.router"),
		metrics:  metrics,
		sem:      make(chan struct{}, concurrency),
	}
}

// Register attaches a handler for a subject.
func (r *Router) Register(subject string, handler events.Handler) {
	r.handlers[subject] = r.wrapHandler(subject, handler)
}

// UsePreHandler registers a callback executed before each event handler.
func (r *Router) UsePreHandler(fn func(context.Context, *events.Event) error) {
	if fn == nil {
		return
	}
	r.pre = append(r.pre, fn)
}

// Start subscribes all registered handlers.
func (r *Router) Start(ctx context.Context) error {
	for subject, handler := range r.handlers {
		if err := r.bus.Subscribe(ctx, subject, "", handler); err != nil {
			return err
		}
	}
	r.logger.Info("event router started", zap.Int("handler_count", len(r.handlers)), zap.String("group", r.group))
	return nil
}

func (r *Router) wrapHandler(subject string, h events.Handler) events.Handler {
	return func(ctx context.Context, evt *events.Event) error {
		if err := r.acquire(ctx); err != nil {
			return err
		}
		defer r.release()

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

		tenant := ""
		eventID := ""
		if evt != nil {
			tenant = evt.GetTenantId()
			eventID = evt.GetId()
		}
		start := time.Now()
		log := ctxlogger.FromContext(ctx).With(
			zap.String("handler_name", subject),
			zap.String("tenant_id", tenant),
			zap.String("event_id", eventID),
		)
		for _, pre := range r.pre {
			if err := pre(ctx, evt); err != nil {
				return err
			}
		}
		if tenant != "" {
			span.SetAttributes(attribute.String("event.tenant_id", tenant))
		}
		if eventID != "" {
			span.SetAttributes(attribute.String("event.id", eventID))
		}

		err := h(ctx, evt)
		status := "success"
		if err != nil {
			status = "error"
			log.Error("handler failed", zap.Error(err))
		} else {
			log.Info("handler completed")
		}
		duration := time.Since(start)
		if r.metrics != nil {
			r.metrics.RecordHandler(subject, tenant, status, duration)
		}
		return err
	}
}

func (r *Router) acquire(ctx context.Context) error {
	if r.sem == nil {
		return nil
	}
	select {
	case r.sem <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Router) release() {
	if r.sem == nil {
		return
	}
	select {
	case <-r.sem:
	default:
	}
}

func handlerConcurrency() int {
	raw := os.Getenv("HANDLER_CONCURRENCY")
	if raw == "" {
		return runtime.NumCPU()
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val < 1 {
		return runtime.NumCPU()
	}
	return val
}
