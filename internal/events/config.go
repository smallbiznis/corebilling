package events

import (
	"os"
	"strings"
)

// EventBusProvider enumerates supported backends.
type EventBusProvider string

const (
	ProviderNATS  EventBusProvider = "nats"
	ProviderKafka EventBusProvider = "kafka"
	ProviderNoop  EventBusProvider = "noop"
)

// EventBusConfig holds provider-specific configuration.
type EventBusConfig struct {
	Provider EventBusProvider

	// NATS
	NATSURL      string
	NATSUsername string
	NATSPassword string
	NATSStream   string

	// Kafka/Confluent
	KafkaBrokers      []string
	KafkaSASLUsername string
	KafkaSASLPassword string
	KafkaGroupID      string
}

// NewEventBusConfig builds configuration from environment variables.
func NewEventBusConfig() EventBusConfig {
	provider := EventBusProvider(strings.ToLower(getenv("EVENT_BUS_PROVIDER", "noop")))
	return EventBusConfig{
		Provider:          provider,
		NATSURL:           getenv("NATS_URL", "nats://localhost"),
		NATSUsername:      getenv("NATS_USERNAME", "natsuser"),
		NATSPassword:      getenv("NATS_PASSWORD", "natspassword"),
		NATSStream:        getenv("NATS_STREAM", "corebilling"),
		KafkaBrokers:      parseCSV(getenv("KAFKA_BROKERS", "")),
		KafkaSASLUsername: getenv("KAFKA_SASL_USERNAME", ""),
		KafkaSASLPassword: getenv("KAFKA_SASL_PASSWORD", ""),
		KafkaGroupID:      getenv("KAFKA_GROUP_ID", "corebilling"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseCSV(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
