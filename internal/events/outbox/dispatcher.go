package outbox

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/telemetry"
)

// Dispatcher reads from the outbox table and publishes events via the configured bus.
type Dispatcher struct {
	repo        OutboxRepository
	bus         events.Publisher
	logger      *zap.Logger
	baseBackoff time.Duration
	maxBackoff  time.Duration
	limit       int32
	metrics     *telemetry.Metrics
}

// NewDispatcher constructs a dispatcher with sane defaults.
func NewDispatcher(repo OutboxRepository, bus events.Publisher, logger *zap.Logger, metrics *telemetry.Metrics) *Dispatcher {
	return &Dispatcher{
		repo:        repo,
		bus:         bus,
		logger:      logger,
		baseBackoff: time.Second,
		maxBackoff:  5 * time.Minute,
		limit:       100,
		metrics:     metrics,
	}
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
	start := time.Now()
	eventsOut, err := d.repo.FetchPendingEvents(ctx, d.limit, now)
	if err != nil {
		if d.metrics != nil {
			d.metrics.RecordOutboxBatch("error", 0, time.Since(start))
		}
		d.logger.Error("failed to fetch outbox events", zap.Error(err))
		return
	}
	if d.metrics != nil {
		d.metrics.SetOutboxBacklog(float64(len(eventsOut)))
	}
	if len(eventsOut) == 0 {
		if d.metrics != nil {
			d.metrics.RecordOutboxBatch("empty", 0, time.Since(start))
		}
		return
	}

	envelopes := make([]events.EventEnvelope, 0, len(eventsOut))
	for _, evt := range eventsOut {
		env := events.NewEnvelope(evt.Event)
		if env.Subject == "" {
			env.Subject = evt.Subject
		}
		if env.TenantID == "" {
			env.TenantID = evt.TenantID
		}
		envelopes = append(envelopes, env)
	}

	if err := d.bus.Publish(ctx, envelopes...); err != nil {
		if d.metrics != nil {
			d.metrics.RecordOutboxBatch("error", len(eventsOut), time.Since(start))
		}
		for _, evt := range eventsOut {
			d.handleFailure(ctx, evt, err)
		}
		return
	}

	if d.metrics != nil {
		d.metrics.RecordOutboxBatch("success", len(eventsOut), time.Since(start))
	}

	for _, evt := range eventsOut {
		if err := d.repo.MarkDispatched(ctx, evt.ID); err != nil {
			d.logger.Error("failed to mark dispatched", zap.Error(err), zap.String("id", evt.ID))
		}
	}
}

// Dispatch processes a single batch of pending events. This helper is exported for testing.
func (d *Dispatcher) Dispatch(ctx context.Context) {
	d.processBatch(ctx)
}

func (d *Dispatcher) handleFailure(ctx context.Context, evt OutboxEvent, err error) {
	retry := evt.RetryCount + 1
	log := d.logger.With(
		zap.String("event_id", evt.ID),
		zap.String("tenant_id", evt.TenantID),
		zap.String("subject", evt.Subject),
	)
	if ShouldMoveToDLQ(retry) {
		if dlqErr := d.repo.MoveToDeadLetter(ctx, evt.ID, err.Error()); dlqErr != nil {
			log.Error("failed to move to DLQ", zap.Error(dlqErr))
		}
		log.Error("event moved to DLQ", zap.Error(err))
		return
	}

	next := ComputeNextAttempt(retry, d.baseBackoff, d.maxBackoff)
	if markErr := d.repo.MarkFailed(ctx, evt.ID, next, err.Error(), retry); markErr != nil {
		log.Error("failed to mark retry", zap.Error(markErr))
	}
}
