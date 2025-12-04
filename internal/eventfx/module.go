package eventfx

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/eventbus"
	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	"github.com/smallbiznis/corebilling/internal/events/router"
)

// Module wires event bus, outbox, and router into the Fx graph.
var Module = fx.Options(
	fx.Provide(
		events.NewEventBusConfig,
		eventbus.NewBus,
		outbox.NewRepository,
		func(repo outbox.OutboxRepository, bus events.Bus, logger *zap.Logger) *outbox.Dispatcher {
			return outbox.NewDispatcher(repo, bus, logger)
		},
		func(bus events.Bus, logger *zap.Logger, cfg events.EventBusConfig) *router.Router {
			group := cfg.KafkaGroupID
			if group == "" {
				group = "corebilling"
			}
			return router.NewRouter(bus, logger, group)
		},
		func() *outbox.IdempotencyTracker {
			tracker := outbox.NewIdempotencyTracker(10 * time.Minute)
			return tracker
		},
	),
	fx.Invoke(
		startDispatcher,
		manageBusLifecycle,
		startIdempotency,
	),
)

func startDispatcher(lc fx.Lifecycle, dispatcher *outbox.Dispatcher, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go dispatcher.Run(ctx)
			logger.Info("outbox dispatcher started")
			return nil
		},
	})
}

func manageBusLifecycle(lc fx.Lifecycle, bus events.Bus, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("closing event bus")
			return bus.Close()
		},
	})
}

func startIdempotency(lc fx.Lifecycle, tracker *outbox.IdempotencyTracker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			tracker.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			tracker.Stop()
			return nil
		},
	})
}
