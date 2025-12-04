package noop

import (
	"context"

	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
)

// NoopBus implements Bus with logging only.
type NoopBus struct {
	logger *zap.Logger
}

// NewNoopBus creates a new noop bus instance.
func NewNoopBus(logger *zap.Logger) *NoopBus {
	return &NoopBus{logger: logger}
}

// Publish logs the event instead of sending it.
func (b *NoopBus) Publish(ctx context.Context, subject string, payload []byte) error {
	b.logger.Info("noop publish", zap.String("subject", subject), zap.Int("payload_bytes", len(payload)))
	return nil
}

// Subscribe logs the subscription and drops messages.
func (b *NoopBus) Subscribe(ctx context.Context, subject string, group string, handler events.Handler) error {
	b.logger.Info("noop subscribe", zap.String("subject", subject), zap.String("group", group))
	return nil
}

// Close performs no cleanup.
func (b *NoopBus) Close() error { return nil }

var _ events.Bus = (*NoopBus)(nil)
