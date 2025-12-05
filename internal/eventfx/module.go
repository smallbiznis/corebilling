package eventfx

import (
	"context"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/eventbus"
	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/handler"
	"github.com/smallbiznis/corebilling/internal/events/handler/providers"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	"github.com/smallbiznis/corebilling/internal/events/router"
	"github.com/smallbiznis/corebilling/internal/telemetry"
)

// Module wires event bus, outbox, and router into the Fx graph.
var Module = fx.Options(
	providers.Module,
	fx.Provide(
		events.NewEventBusConfig,
		eventbus.NewBus,
		providePublisher,
		fx.Annotate(outbox.NewRepository, fx.As(new(outbox.OutboxRepository)), fx.As(new(outbox.DeadLetterRepository))),
		func(repo outbox.OutboxRepository, bus events.Bus, logger *zap.Logger, metrics *telemetry.Metrics) *outbox.Dispatcher {
			return outbox.NewDispatcher(repo, bus, logger, metrics)
		},
		outbox.NewDLQService,
		func(bus events.Bus, logger *zap.Logger, cfg events.EventBusConfig, metrics *telemetry.Metrics) *router.Router {
			group := cfg.KafkaGroupID
			if group == "" {
				group = "corebilling"
			}
			return router.NewRouter(bus, logger, group, metrics)
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
		registerHandlers,
		startRouter,
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
func registerHandlers(router *router.Router, group handler.HandlerGroup) {
	for _, h := range group.Handlers {
		router.Register(h.Subject(), h.Handle)
	}
}

func providePublisher(bus events.Bus) events.Publisher {
	return bus
}

func startRouter(lc fx.Lifecycle, router *router.Router, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("starting event router")
			return router.Start(ctx)
		},
	})
}
