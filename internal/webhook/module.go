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
	"subscription.created.v1",
	"subscription.updated.v1",
	"subscription.upgraded.v1",
	"subscription.canceled.v1",
	"subscription.provisioned.v1",
	"subscription.price.updated.v1",
	"subscription.status.changed.v1",
	"usage.reported.v1",
	"usage.ingested.v1",
	"usage.rated.v1",
	"usage.aggregated.v1",
	"usage.status.changed.v1",
	"rating.completed.v1",
	"rating.failed.v1",
	"invoice.created.v1",
	"invoice.opened.v1",
	"invoice.generated.v1",
	"invoice.sent.v1",
	"invoice.paid.v1",
	"invoice.due.v1",
	"invoice.voided.v1",
	"invoice.status.changed.v1",
	"credit.applied.v1",
	"credit.reversed.v1",
	"plan.created.v1",
	"plan.updated.v1",
	"plan.deprecated.v1",
}
