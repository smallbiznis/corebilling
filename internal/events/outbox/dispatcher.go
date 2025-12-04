package outbox

import (
	"context"
	"time"

	"go.uber.org/zap"
)

// Dispatcher reads from the outbox table and publishes events via the configured bus.
type Dispatcher struct {
	repo        OutboxRepository
	bus         Publisher
	logger      *zap.Logger
	baseBackoff time.Duration
	maxBackoff  time.Duration
	limit       int32
}

// Publisher captures the publish capability of an event bus.
type Publisher interface {
	Publish(ctx context.Context, subject string, payload []byte) error
}

// NewDispatcher constructs a dispatcher with sane defaults.
func NewDispatcher(repo OutboxRepository, bus Publisher, logger *zap.Logger) *Dispatcher {
	return &Dispatcher{repo: repo, bus: bus, logger: logger, baseBackoff: time.Second, maxBackoff: 5 * time.Minute, limit: 100}
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
		if err := d.bus.Publish(ctx, evt.Subject, evt.Payload); err != nil {
			d.handleFailure(ctx, evt, err)
			continue
		}
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
