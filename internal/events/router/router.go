package router

import (
	"context"

	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
)

// Router wires subscriptions to handlers across providers.
type Router struct {
	bus      events.Bus
	logger   *zap.Logger
	handlers map[string]events.Handler
	group    string
}

// NewRouter constructs a router with a consumer group identifier.
func NewRouter(bus events.Bus, logger *zap.Logger, group string) *Router {
	return &Router{bus: bus, logger: logger, handlers: make(map[string]events.Handler), group: group}
}

// Register attaches a handler for a subject.
func (r *Router) Register(subject string, handler events.Handler) {
	r.handlers[subject] = handler
}

// Start subscribes all registered handlers.
func (r *Router) Start(ctx context.Context) error {
	for subject, handler := range r.handlers {
		if err := r.bus.Subscribe(ctx, subject, r.group, handler); err != nil {
			return err
		}
	}
	r.logger.Info("event router started", zap.Int("handler_count", len(r.handlers)), zap.String("group", r.group))
	return nil
}
