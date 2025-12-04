package events

import "context"

// Handler represents a consumer callback for an event.
type Handler func(ctx context.Context, evt *Event) error

// Bus defines a provider-agnostic event bus contract.
type Bus interface {
	Publish(ctx context.Context, subject string, payload []byte) error
	Subscribe(ctx context.Context, subject string, group string, handler Handler) error
	Close() error
}
