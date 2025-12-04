package bus

import "context"

// EventBus defines a simple publish interface used by dispatchers.
type EventBus interface {
	Publish(ctx context.Context, subject string, data []byte) error
}
