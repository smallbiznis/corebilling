package webhook

import (
	"context"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/events/router"
	"github.com/smallbiznis/corebilling/internal/webhook/domain"
	"github.com/smallbiznis/corebilling/internal/webhook/repository"
	reposqlc "github.com/smallbiznis/corebilling/internal/webhook/repository/sqlc"
)

var Module = fx.Options(
	fx.Provide(
		LoadConfig,
		repository.NewRepository,
		NewService,
		reposqlc.NewRepository,
		domain.NewService,
		newHTTPClient,
		NewWorker,
	),
	fx.Invoke(
		startWorker,
		registerWebhookHandlers,
	),
	ModuleGRPC,
)

var ModuleGRPC = fx.Invoke(RegisterGRPC)

func newHTTPClient(cfg Config) *http.Client {
	return &http.Client{Timeout: cfg.HTTPTimeout}
}

func startWorker(lc fx.Lifecycle, worker *Worker, logger *zap.Logger) {
	var cancel context.CancelFunc
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			runCtx, c := context.WithCancel(ctx)
			cancel = c
			go worker.Run(runCtx)
			logger.Info("webhook delivery worker started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if cancel != nil {
				cancel()
			}
			logger.Info("stopping webhook delivery worker")
			return nil
		},
	})
}

func registerWebhookHandlers(router *router.Router, svc *Service, logger *zap.Logger) {
	log := logger.Named("webhook.router")
	for _, subject := range webhookSubjects {
		subject := subject
		router.Register(subject, func(ctx context.Context, evt *events.Event) error {
			env := events.NewEnvelope(evt)
			env.Subject = subject
			ctx, err := env.Prepare(ctx)
			if err != nil {
				log.Error("failed to prepare webhook envelope", zap.String("subject", subject), zap.Error(err))
				return err
			}
			if err := svc.DispatchForEvent(ctx, env); err != nil {
				log.Error("webhook dispatch failed", zap.String("subject", subject), zap.String("event_id", env.Event.GetId()), zap.Error(err))
				return err
			}
			return nil
		})
	}
}

var webhookSubjects = []string{
	"subscription.created",
	"subscription.updated",
	"subscription.upgraded",
	"subscription.canceled",
	"subscription.provisioned",
	"subscription.price.updated",
	"subscription.status.changed",
	"usage.reported",
	"usage.ingested",
	"usage.rated",
	"usage.aggregated",
	"usage.status.changed",
	"rating.completed",
	"rating.failed",
	"invoice.created",
	"invoice.opened",
	"invoice.generated",
	"invoice.sent",
	"invoice.paid",
	"invoice.due",
	"invoice.voided",
	"invoice.status.changed",
	"credit.applied",
	"credit.reversed",
	"plan.created",
	"plan.updated",
	"plan.deprecated",
}
