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
func (b *KafkaBus) Publish(ctx context.Context, subject string, payload []byte) error {
	var key []byte
	if evt, err := events.UnmarshalEvent(payload); err == nil && evt.GetTenantId() != "" {
		key = []byte(evt.GetTenantId())
	}

	ctx, span := b.tracer.Start(ctx, "event.publish", trace.WithAttributes(
		attribute.String("event.subject", subject),
		attribute.String("event.provider", "kafka"),
	))
	defer span.End()

	msg := &ckafka.Message{TopicPartition: ckafka.TopicPartition{Topic: &subject, Partition: ckafka.PartitionAny}, Value: payload}
	if len(key) > 0 {
		msg.Key = key
	}

	if err := b.producer.Produce(msg, nil); err != nil {
		span.RecordError(err)
		return err
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

	attrs := []attribute.KeyValue{
		attribute.String("event.subject", evt.GetSubject()),
		attribute.String("event.provider", "kafka"),
	}
	if tenant := evt.GetTenantId(); tenant != "" {
		attrs = append(attrs, attribute.String("event.tenant_id", tenant))
	}

	ctx, span := b.tracer.Start(context.Background(), "event.consume", trace.WithAttributes(attrs...))
	defer span.End()

	if err := handler(ctx, evt); err != nil {
		span.RecordError(err)
		b.logger.Error("handler failed", zap.Error(err), zap.String("topic", topic))
		return
	}

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
