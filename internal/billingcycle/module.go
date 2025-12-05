package billingcycle

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/billingcycle/repository"
	"github.com/smallbiznis/corebilling/internal/invoice_engine/domain"
)

// Module wires billing cycle services and scheduler.
var Module = fx.Options(
	fx.Provide(repository.NewRepository),
	fx.Provide(func(engine *domain.Service) InvoiceGenerator {
		return &engineAdapter{engine: engine}
	}),
	fx.Provide(NewService),
	fx.Provide(NewScheduler),
	fx.Invoke(startScheduler),
)

type engineAdapter struct {
	engine *domain.Service
}

func (e *engineAdapter) GenerateForTenant(ctx context.Context, tenantID string) error {
	// Placeholder implementation; actual invoice generation would enumerate subscriptions.
	return nil
}

func startScheduler(lc fx.Lifecycle, scheduler *Scheduler, logger *zap.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go scheduler.Run(ctx)
			logger.Info("billing cycle scheduler started")
			return nil
		},
	})
}
