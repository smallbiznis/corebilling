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
	"github.com/smallbiznis/corebilling/internal/telemetry/correlation"
	"google.golang.org/protobuf/types/known/structpb"
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
	ctx, cid := correlation.EnsureCorrelationID(ctx)
	ctx, span := b.tracer.Start(ctx, "nats.publish", trace.WithSpanKind(trace.SpanKindProducer), trace.WithAttributes(
		attribute.String("messaging.system", "nats"),
		attribute.String("messaging.destination_kind", "topic"),
		attribute.String("messaging.destination", subject),
		attribute.String("correlation_id", cid),
		attribute.String("trace_id", span.SpanContext().TraceID().String()),
	))
	defer span.End()

	evt, err := events.UnmarshalEvent(payload)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if evt != nil {
		if cid != "" {
			if evt.Metadata == nil {
				evt.Metadata = &structpb.Struct{Fields: map[string]*structpb.Value{}}
			}
			if evt.Metadata.Fields == nil {
				evt.Metadata.Fields = map[string]*structpb.Value{}
			}
			evt.Metadata.Fields["correlation_id"] = structpb.NewStringValue(cid)
		}
		correlation.InjectTraceIntoEvent(evt, span)
		if updated, marshalErr := events.MarshalEvent(evt); marshalErr == nil {
			payload = updated
			if val, ok := evt.Metadata.Fields["correlation_id"]; ok {
				cid = val.GetStringValue()
			}
		} else {
			span.RecordError(marshalErr)
			return marshalErr
		}
	}

	hdr := nats.Header{}
	if cid != "" {
		hdr.Set("correlation-id", cid)
	}
	hdr.Set("trace-id", span.SpanContext().TraceID().String())
	hdr.Set("span-id", span.SpanContext().SpanID().String())

	msg := &nats.Msg{Subject: subject, Data: payload, Header: hdr}
	if _, err := b.js.PublishMsg(msg); err != nil {
		span.RecordError(err)
		return err
	}
	return nil
}

// Subscribe registers a queue subscription for the subject.
func (b *NATSBus) Subscribe(ctx context.Context, subject, group string, handler events.Handler) error {
	_, err := b.js.QueueSubscribe(subject, group, func(msg *nats.Msg) {
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

	ctx := correlation.InjectCorrelationID(context.Background(), cid)
	ctx = correlation.ContextWithRemoteSpan(ctx, tid, sid)

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
