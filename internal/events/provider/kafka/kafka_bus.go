package kafka

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/log/ctxlogger"
	"github.com/smallbiznis/corebilling/internal/telemetry/correlation"
)

// KafkaBus implements events.Bus using Confluent's Go client.
type KafkaBus struct {
	producer *ckafka.Producer
	consumer *ckafka.Consumer
	logger   *zap.Logger
	cfg      events.EventBusConfig
	tracer   trace.Tracer

	mu       sync.Mutex
	handlers map[string]events.Handler
	cancel   context.CancelFunc
	running  bool
}

// NewKafkaBus constructs a Kafka-backed bus.
func NewKafkaBus(cfg events.EventBusConfig, logger *zap.Logger) (events.Bus, error) {
	if len(cfg.KafkaBrokers) == 0 {
		return nil, fmt.Errorf("kafka brokers required")
	}

	commonCfg := ckafka.ConfigMap{
		"bootstrap.servers": strings.Join(cfg.KafkaBrokers, ","),
	}
	if cfg.KafkaSASLUsername != "" && cfg.KafkaSASLPassword != "" {
		commonCfg["security.protocol"] = "SASL_SSL"
		commonCfg["sasl.mechanisms"] = "PLAIN"
		commonCfg["sasl.username"] = cfg.KafkaSASLUsername
		commonCfg["sasl.password"] = cfg.KafkaSASLPassword
	}

	prodCfg := commonCfg
	producer, err := ckafka.NewProducer(&prodCfg)
	if err != nil {
		return nil, fmt.Errorf("kafka producer: %w", err)
	}

	consCfg := commonCfg
	consCfg["group.id"] = cfg.KafkaGroupID
	consCfg["auto.offset.reset"] = "earliest"
	consCfg["enable.auto.commit"] = false

	consumer, err := ckafka.NewConsumer(&consCfg)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: %w", err)
	}

	return &KafkaBus{
		producer: producer,
		consumer: consumer,
		logger:   logger,
		cfg:      cfg,
		tracer:   otel.Tracer("events.kafka"),
		handlers: make(map[string]events.Handler),
	}, nil
}

// Publish sends an event to Kafka.
func (b *KafkaBus) Publish(ctx context.Context, envelopes ...events.EventEnvelope) error {
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
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination_kind", "topic"),
			attribute.String("messaging.destination", env.Subject),
		}
		if env.TenantID != "" {
			attrs = append(attrs, attribute.String("event.tenant_id", env.TenantID))
		}
		if env.CorrelationID != "" {
			attrs = append(attrs, attribute.String("correlation_id", env.CorrelationID))
		}

		publishCtx, span := b.tracer.Start(publishCtx, "kafka.publish", trace.WithSpanKind(trace.SpanKindProducer), trace.WithAttributes(attrs...))
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

		msg := &ckafka.Message{TopicPartition: ckafka.TopicPartition{Topic: &env.Subject, Partition: ckafka.PartitionAny}, Value: env.Payload}
		if env.TenantID != "" {
			msg.Key = []byte(env.TenantID)
		}
		headers := []ckafka.Header{
			{Key: "trace-id", Value: []byte(span.SpanContext().TraceID().String())},
			{Key: "span-id", Value: []byte(span.SpanContext().SpanID().String())},
		}
		if env.CorrelationID != "" {
			headers = append(headers, ckafka.Header{Key: "correlation-id", Value: []byte(env.CorrelationID)})
		}
		if env.CausationID != "" {
			headers = append(headers, ckafka.Header{Key: "causation-id", Value: []byte(env.CausationID)})
		}
		if env.TenantID != "" {
			headers = append(headers, ckafka.Header{Key: "tenant-id", Value: []byte(env.TenantID)})
		}
		msg.Headers = headers

		if err := b.producer.Produce(msg, nil); err != nil {
			span.RecordError(err)
			span.End()
			return err
		}
		span.End()
	}
	return nil
}

// Subscribe registers a handler for the given subject/topic.
func (b *KafkaBus) Subscribe(ctx context.Context, subject string, group string, handler events.Handler) error {
	b.mu.Lock()
	b.handlers[subject] = handler

	topics := make([]string, 0, len(b.handlers))
	for topic := range b.handlers {
		topics = append(topics, topic)
	}
	err := b.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		b.mu.Unlock()
		return fmt.Errorf("subscribe topics: %w", err)
	}

	if !b.running {
		runCtx, cancel := context.WithCancel(context.Background())
		b.cancel = cancel
		b.running = true
		go b.poll(runCtx)
	}
	b.mu.Unlock()

	b.logger.Info("kafka subscription established", zap.String("subject", subject), zap.String("group", b.cfg.KafkaGroupID))
	return nil
}

func (b *KafkaBus) poll(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ev := b.consumer.Poll(500)
		if ev == nil {
			continue
		}

		switch msg := ev.(type) {
		case *ckafka.Message:
			b.handleMessage(msg)
		case ckafka.Error:
			b.logger.Error("kafka error", zap.Error(msg))
		default:
		}
	}
}

func (b *KafkaBus) handleMessage(msg *ckafka.Message) {
	topic := ""
	if msg.TopicPartition.Topic != nil {
		topic = *msg.TopicPartition.Topic
	}

	b.mu.Lock()
	handler, ok := b.handlers[topic]
	b.mu.Unlock()
	if !ok {
		b.logger.Warn("no handler for topic", zap.String("topic", topic))
		return
	}

	evt, err := events.UnmarshalEvent(msg.Value)
	if err != nil {
		b.logger.Error("failed to decode event", zap.Error(err))
		return
	}

	var cid, tid, sid string
	for _, h := range msg.Headers {
		switch h.Key {
		case "correlation-id":
			cid = string(h.Value)
		case "trace-id":
			tid = string(h.Value)
		case "span-id":
			sid = string(h.Value)
		}
	}

	ctx := correlation.ContextWithCorrelationID(context.Background(), cid)
	ctx = correlation.ContextWithRemoteSpan(ctx, tid, sid)
	ctx = ctxlogger.ContextWithEventSubject(ctx, evt.GetSubject())

	log := ctxlogger.FromContext(ctx)

	attrs := []attribute.KeyValue{
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.kafka.topic", topic),
		attribute.Int64("messaging.kafka.partition", int64(msg.TopicPartition.Partition)),
		attribute.Int64("messaging.kafka.offset", int64(msg.TopicPartition.Offset)),
		attribute.String("correlation_id", cid),
		attribute.String("trace_id", tid),
	}
	if tenant := evt.GetTenantId(); tenant != "" {
		attrs = append(attrs, attribute.String("event.tenant_id", tenant))
	}

	ctx, span := b.tracer.Start(ctx, "kafka.consume", trace.WithSpanKind(trace.SpanKindConsumer), trace.WithAttributes(attrs...))
	defer span.End()
	correlation.InjectTraceIntoEvent(evt, span)

	log.Info("event.consume.start", zap.String("subject", topic), zap.String("correlation_id", cid))
	if err := handler(ctx, evt); err != nil {
		span.RecordError(err)
		b.logger.Error("handler failed", zap.Error(err), zap.String("topic", topic))
		return
	}

	log.Info("event.consume.finish")

	if _, err := b.consumer.CommitMessage(msg); err != nil {
		b.logger.Warn("commit failed", zap.Error(err))
	}
}

// Close shuts down producer and consumer.
func (b *KafkaBus) Close() error {
	if b.cancel != nil {
		b.cancel()
	}
	if b.consumer != nil {
		_ = b.consumer.Close()
	}
	if b.producer != nil {
		b.producer.Flush(int(5 * time.Second / time.Millisecond))
		b.producer.Close()
	}
	return nil
}

var _ events.Bus = (*KafkaBus)(nil)
