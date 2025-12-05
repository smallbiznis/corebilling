package bus

import (
	"context"

	"github.com/smallbiznis/corebilling/internal/events"
)

// EventBus defines a simple publish interface used by dispatchers.
type EventBus interface {
	Publish(ctx context.Context, envelopes ...events.EventEnvelope) error
}
