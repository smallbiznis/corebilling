package handler

import (
	"context"

	"github.com/smallbiznis/corebilling/internal/events"
	"go.uber.org/fx"
)

// EventHandler reacts to a specific event subject.
type EventHandler interface {
	Subject() string
	Handle(ctx context.Context, evt *events.Event) error
}

// HandlerOut exposes an EventHandler for fx grouping.
type HandlerOut struct {
	fx.Out
	Handler EventHandler `group:"event_handlers"`
}

// HandlerGroup bundles all registered handlers.
type HandlerGroup struct {
	fx.In
	Handlers []EventHandler `group:"event_handlers"`
}
