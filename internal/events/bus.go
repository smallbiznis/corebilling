package events

import (
	"context"

	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
)

// Handler represents a consumer callback for an event.
type Handler func(ctx context.Context, evt *eventv1.Event) error

// Bus defines a provider-agnostic event bus contract.
type Bus interface {
	Publish(ctx context.Context, subject string, payload []byte) error
	Subscribe(ctx context.Context, subject string, group string, handler Handler) error
	Close() error
}
