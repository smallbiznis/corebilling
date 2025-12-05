package events

import (
	"context"

	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
)

// Handler represents a consumer callback for an event.
type Handler func(ctx context.Context, evt *eventv1.Event) error

// Publisher publishes events in batches.
type Publisher interface {
	Publish(ctx context.Context, envelopes ...EventEnvelope) error
}

// Bus defines a provider-agnostic event bus contract.
type Bus interface {
	Publisher
	Subscribe(ctx context.Context, subject string, group string, handler Handler) error
	Close() error
}
