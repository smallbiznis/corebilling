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
	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
	"github.com/smallbiznis/corebilling/internal/telemetry/correlation"
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
	subjects := []string{
		"subscription.>",
		"invoice.>",
		"usage.>",
		"rating.>",
		"credit.>",
		"plan.>",
	}
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
func (b *NATSBus) Publish(ctx context.Context, envelopes ...events.EventEnvelope) error {
	if len(envelopes) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(envelopes))
	for _, env := range envelopes {
		publishCtx, err := env.Prepare(ctx)
		if err != nil {
			return err
		}
		if env.Event == nil || env.Subject == "" {
			continue
		}
		if _, ok := seen[env.Event.Id]; ok {
			continue
		}
		seen[env.Event.Id] = struct{}{}

		attrs := []attribute.KeyValue{
			attribute.String("messaging.system", "nats"),
			attribute.String("messaging.destination_kind", "topic"),
			attribute.String("messaging.destination", env.Subject),
		}
		if env.TenantID != "" {
			attrs = append(attrs, attribute.String("event.tenant_id", env.TenantID))
		}
		if env.CorrelationID != "" {
			attrs = append(attrs, attribute.String("correlation_id", env.CorrelationID))
		}

		publishCtx, span := b.tracer.Start(publishCtx, "nats.publish", trace.WithSpanKind(trace.SpanKindProducer), trace.WithAttributes(attrs...))
		correlation.InjectTraceIntoEvent(env.Event, span)
		payload, marshalErr := events.MarshalEvent(env.Event)
		if marshalErr != nil {
			span.RecordError(marshalErr)
			span.End()
			return marshalErr
		}
		env.Payload = payload

		log := ctxlogger.FromContext(publishCtx).With(zap.String("subject", env.Subject), zap.String("event_id", env.Event.Id))
		if env.Event.Metadata != nil {
			log = log.With(zap.Any("metadata", env.Event.Metadata.AsMap()))
		}
		log.Info("event.publish")

		hdr := nats.Header{}
		if env.CorrelationID != "" {
			hdr.Set("correlation-id", env.CorrelationID)
		}
		if env.CausationID != "" {
			hdr.Set("causation-id", env.CausationID)
		}
		if env.TenantID != "" {
			hdr.Set("tenant-id", env.TenantID)
		}
		hdr.Set("trace-id", span.SpanContext().TraceID().String())
		hdr.Set("span-id", span.SpanContext().SpanID().String())

		msg := &nats.Msg{Subject: env.Subject, Data: env.Payload, Header: hdr}
		if _, err := b.js.PublishMsg(msg); err != nil {
			span.RecordError(err)
			span.End()
			return err
		}
		span.End()
	}
	return nil
}

// Subscribe registers a queue subscription for the subject.
func (b *NATSBus) Subscribe(ctx context.Context, subject, group string, handler events.Handler) error {
	_, err := b.js.Subscribe(subject, func(msg *nats.Msg) {
		go b.handleMessage(msg, handler, group)
	})
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}
	b.logger.Info("nats subscription established", zap.String("subject", subject), zap.String("group", group))
	return nil
}

func (b *NATSBus) handleMessage(msg *nats.Msg, handler events.Handler, group string) {
	evt, err := events.UnmarshalEvent(msg.Data)
	if err != nil {
		b.logger.Error("failed to decode event", zap.Error(err))
		_ = msg.Nak()
		return
	}

	cid := msg.Header.Get("correlation-id")
	tid := msg.Header.Get("trace-id")
	sid := msg.Header.Get("span-id")

	ctx := correlation.ContextWithCorrelationID(context.Background(), cid)
	ctx = correlation.ContextWithRemoteSpan(ctx, tid, sid)
	ctx = ctxlogger.ContextWithEventSubject(ctx, evt.GetSubject())

	log := ctxlogger.FromContext(ctx)

	attrs := []attribute.KeyValue{
		attribute.String("messaging.system", "nats"),
		attribute.String("messaging.operation", "process"),
		attribute.String("messaging.destination", evt.GetSubject()),
		attribute.String("correlation_id", cid),
		attribute.String("trace_id", tid),
		attribute.String("messaging.consumer_group", group),
	}
	if tenant := evt.GetTenantId(); tenant != "" {
		attrs = append(attrs, attribute.String("event.tenant_id", tenant))
	}

	ctx, span := b.tracer.Start(ctx, "nats.consume", trace.WithSpanKind(trace.SpanKindConsumer), trace.WithAttributes(attrs...))
	defer span.End()
	correlation.InjectTraceIntoEvent(evt, span)

	log.Info("event.consume.start", zap.String("subject", evt.GetSubject()), zap.String("correlation_id", cid))
	if err := handler(ctx, evt); err != nil {
		span.RecordError(err)
		b.logger.Error("handler failed", zap.Error(err), zap.String("subject", evt.GetSubject()))
		_ = msg.Nak()
		return
	}

	log.Info("event.consume.finish")

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
