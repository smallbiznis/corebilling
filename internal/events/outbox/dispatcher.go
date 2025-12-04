package outbox

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
	"github.com/smallbiznis/corebilling/internal/telemetry/correlation"
	"google.golang.org/protobuf/types/known/structpb"
)

// Dispatcher reads from the outbox table and publishes events via the configured bus.
type Dispatcher struct {
	repo        OutboxRepository
	bus         Publisher
	logger      *zap.Logger
	baseBackoff time.Duration
	maxBackoff  time.Duration
	limit       int32
	tracer      trace.Tracer
}

// Publisher captures the publish capability of an event bus.
type Publisher interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}

// NewDispatcher constructs a dispatcher with sane defaults.
func NewDispatcher(repo OutboxRepository, bus Publisher, logger *zap.Logger) *Dispatcher {
	return &Dispatcher{repo: repo, bus: bus, logger: logger, baseBackoff: time.Second, maxBackoff: 5 * time.Minute, limit: 100, tracer: otel.Tracer("events.outbox.dispatcher")}
}

// Run continuously dispatches pending events until context cancellation.
func (d *Dispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.processBatch(ctx)
		}
	}
}

func (d *Dispatcher) processBatch(ctx context.Context) {
	now := time.Now().UTC()
	eventsOut, err := d.repo.FetchPendingEvents(ctx, d.limit, now)
	if err != nil {
		d.logger.Error("failed to fetch outbox events", zap.Error(err))
		return
	}

	for _, evt := range eventsOut {
		corrID := extractCorrelationFromEvent(evt.Event)
		eventCtx := ctxlogger.ContextWithEventSubject(ctx, evt.Subject)
		if corrID != "" {
			eventCtx = correlation.ContextWithCorrelationID(eventCtx, corrID)
		}

		spanAttrs := []attribute.KeyValue{
			attribute.String("event.subject", evt.Subject),
			attribute.String("outbox.id", evt.ID),
		}
		if corrID != "" {
			spanAttrs = append(spanAttrs, attribute.String("event.correlation_id", corrID))
		}

		eventCtx, span := d.tracer.Start(eventCtx, "outbox.dispatch", trace.WithAttributes(spanAttrs...))
		if evt.Event != nil {
			if corrID != "" {
				if evt.Event.Metadata == nil {
					evt.Event.Metadata = &structpb.Struct{Fields: map[string]*structpb.Value{}}
				}
				if evt.Event.Metadata.Fields == nil {
					evt.Event.Metadata.Fields = map[string]*structpb.Value{}
				}
				evt.Event.Metadata.Fields["correlation_id"] = structpb.NewStringValue(corrID)
			}
			correlation.InjectTraceIntoEvent(evt.Event, span)
			if updated, err := events.MarshalEvent(evt.Event); err == nil {
				evt.Payload = updated
			}
		}
		log := ctxlogger.FromContext(eventCtx)
		log.Info("outbox.dispatch", zap.String("subject", evt.Subject))

		if err := d.bus.Publish(eventCtx, evt.Subject, evt.Payload); err != nil {
			span.RecordError(err)
			d.handleFailure(ctx, evt, err)
			span.End()
			continue
		}
		span.End()
		if err := d.repo.MarkDispatched(ctx, evt.ID); err != nil {
			d.logger.Error("failed to mark dispatched", zap.Error(err), zap.String("id", evt.ID))
		}
	}
}

func (d *Dispatcher) handleFailure(ctx context.Context, evt OutboxEvent, err error) {
	retry := evt.RetryCount + 1
	if ShouldMoveToDLQ(retry) {
		if dlqErr := d.repo.MoveToDeadLetter(ctx, evt.ID, err.Error()); dlqErr != nil {
			d.logger.Error("failed to move to DLQ", zap.Error(dlqErr), zap.String("id", evt.ID))
		}
		d.logger.Error("event moved to DLQ", zap.String("id", evt.ID), zap.Error(err))
		return
	}

	next := ComputeNextAttempt(retry, d.baseBackoff, d.maxBackoff)
	if markErr := d.repo.MarkFailed(ctx, evt.ID, next, err.Error(), retry); markErr != nil {
		d.logger.Error("failed to mark retry", zap.Error(markErr), zap.String("id", evt.ID))
	}
}

func extractCorrelationFromEvent(evt *events.Event) string {
	if evt == nil || evt.Metadata == nil {
		return ""
	}

	if val, ok := evt.Metadata.Fields["correlation_id"]; ok {
		return val.GetStringValue()
	}
	return ""
}
