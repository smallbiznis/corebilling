package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
)

// NATSBus implements the events.Bus using NATS JetStream.
type NATSBus struct {
	conn   *nats.Conn
	js     nats.JetStreamContext
	logger *zap.Logger
	tracer trace.Tracer
	stream string
}

// NewNATSBus constructs a JetStream-backed bus.
func NewNATSBus(cfg events.EventBusConfig, logger *zap.Logger) (events.Bus, error) {
	opts := []nats.Option{nats.Name("corebilling-events")}
	if cfg.NATSUsername != "" || cfg.NATSPassword != "" {
		opts = append(opts, nats.UserInfo(cfg.NATSUsername, cfg.NATSPassword))
	}

	conn, err := nats.Connect(cfg.NATSURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}

	js, err := conn.JetStream()
	if err != nil {
		return nil, fmt.Errorf("jetstream: %w", err)
	}

	streamName := cfg.NATSStream
	subjects := []string{"billing.*.*.v1", "subscription.*.*.v1", "customer.*.*.v1"}
	if _, err := js.StreamInfo(streamName); err != nil {
		if err == nats.ErrStreamNotFound {
			if _, err := js.AddStream(&nats.StreamConfig{Name: streamName, Subjects: subjects}); err != nil {
				return nil, fmt.Errorf("add stream: %w", err)
			}
			logger.Info("created nats stream", zap.String("name", streamName))
		} else {
			return nil, fmt.Errorf("stream info: %w", err)
		}
	}

	return &NATSBus{
		conn:   conn,
		js:     js,
		logger: logger,
		tracer: otel.Tracer("events.nats"),
		stream: streamName,
	}, nil
}

// Publish sends the payload to the configured subject.
func (b *NATSBus) Publish(ctx context.Context, subject string, payload []byte) error {
	ctx, span := b.tracer.Start(ctx, "event.publish", trace.WithAttributes(
		attribute.String("event.subject", subject),
		attribute.String("event.provider", "nats"),
	))
	defer span.End()

	if _, err := b.js.Publish(subject, payload); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// Subscribe registers a queue subscription for the subject.
func (b *NATSBus) Subscribe(ctx context.Context, subject, group string, handler events.Handler) error {
	_, err := b.js.QueueSubscribe(subject, group, func(msg *nats.Msg) {
		go b.handleMessage(msg, handler)
	})
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	b.logger.Info("nats subscription established", zap.String("subject", subject), zap.String("group", group))
	return nil
}

func (b *NATSBus) handleMessage(msg *nats.Msg, handler events.Handler) {
	evt, err := events.UnmarshalEvent(msg.Data)
	if err != nil {
		b.logger.Error("failed to decode event", zap.Error(err))
		_ = msg.Nak()
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("event.subject", evt.GetSubject()),
		attribute.String("event.provider", "nats"),
	}
	if tenant := evt.GetTenantId(); tenant != "" {
		attrs = append(attrs, attribute.String("event.tenant_id", tenant))
	}

	ctx, span := b.tracer.Start(context.Background(), "event.consume", trace.WithAttributes(attrs...))
	defer span.End()

	if err := handler(ctx, evt); err != nil {
		span.RecordError(err)
		b.logger.Error("handler failed", zap.Error(err), zap.String("subject", evt.GetSubject()))
		_ = msg.Nak()
		return
	}

	if err := msg.Ack(); err != nil {
		b.logger.Warn("ack failed", zap.Error(err))
	}
}

// Close shuts down the NATS connection.
func (b *NATSBus) Close() error {
	b.conn.Drain()
	b.conn.Close()
	return nil
}

var _ events.Bus = (*NATSBus)(nil)
