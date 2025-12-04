package dispatcher

import (
	"context"
	"time"

	"github.com/smallbiznis/corebilling/internal/events/bus"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
)

// Dispatcher reads pending outbox events and publishes them using an EventBus.
type Dispatcher struct {
	repo outbox.OutboxRepository
	bus  bus.EventBus
}

// NewDispatcher constructs a Dispatcher with provided repository and bus.
func NewDispatcher(r outbox.OutboxRepository, b bus.EventBus) *Dispatcher {
	return &Dispatcher{repo: r, bus: b}
}

// Process fetches up to `limit` pending events and publishes them.
func (d *Dispatcher) Process(ctx context.Context, limit int) error {
	now := time.Now().UTC()
	events, err := d.repo.FetchPendingEvents(ctx, int32(limit), now)
	if err != nil {
		return err
	}
	for _, evt := range events {
		subject := evt.Subject
		payload := evt.Payload
		if evt.Event != nil && evt.Event.Subject != "" {
			subject = evt.Event.Subject
		}
		if err := d.bus.Publish(ctx, subject, payload); err != nil {
			return err
		}
		if err := d.repo.MarkDispatched(ctx, evt.ID); err != nil {
			return err
		}
	}
	return nil
}
