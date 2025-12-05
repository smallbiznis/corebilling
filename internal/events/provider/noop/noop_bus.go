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
func (b *NoopBus) Publish(ctx context.Context, envelopes ...events.EventEnvelope) error {
	for _, env := range envelopes {
		payloadBytes := len(env.Payload)
		if payloadBytes == 0 && env.Event != nil {
			if data, err := events.MarshalEvent(env.Event); err == nil {
				env.Payload = data
				payloadBytes = len(data)
			}
		}
		b.logger.Info("noop publish", zap.String("subject", env.Subject), zap.Int("payload_bytes", payloadBytes))
	}
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
