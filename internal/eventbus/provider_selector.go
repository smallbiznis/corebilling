package eventbus

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	kafkaprovider "github.com/smallbiznis/corebilling/internal/events/provider/kafka"
	natsprovider "github.com/smallbiznis/corebilling/internal/events/provider/nats"
	noopprovider "github.com/smallbiznis/corebilling/internal/events/provider/noop"
)

// NewBus selects and constructs the configured event bus provider.
func NewBus(cfg events.EventBusConfig, logger *zap.Logger) (events.Bus, error) {
	switch cfg.Provider {
	case events.ProviderNATS:
		return natsprovider.NewNATSBus(cfg, logger)
	case events.ProviderKafka:
		return kafkaprovider.NewKafkaBus(cfg, logger)
	case events.ProviderNoop:
		return noopprovider.NewNoopBus(logger), nil
	default:
		return nil, fmt.Errorf("unknown EVENT_BUS_PROVIDER: %s", cfg.Provider)
	}
}
