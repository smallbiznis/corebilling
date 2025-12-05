package quota

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/smallbiznis/corebilling/internal/events/router"
	"github.com/smallbiznis/corebilling/internal/quota/repository"
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
)

// Module wires quota services.
var Module = fx.Options(
	fx.Provide(repository.NewRepository),
	fx.Provide(func() RateLimiter { return NewRateLimiter(10, 100) }),
	fx.Provide(NewService),
	fx.Provide(fx.Annotate(NewHTTPMiddleware, fx.ResultTags(`group:"http_middleware"`))),
	fx.Invoke(attachQuotaChecks),
)

func attachQuotaChecks(r *router.Router, svc *Service, rl RateLimiter, logger *zap.Logger) {
	r.UsePreHandler(func(ctx context.Context, evt *eventv1.Event) error {
		tenant := ""
		if evt != nil {
			tenant = evt.GetTenantId()
		}
		if !rl.Allow(tenant) {
			logger.Warn("rate limit exceeded", zap.String("tenant_id", tenant))
			return ErrQuotaExceeded
		}
		if err := svc.CheckQuotaForEvent(ctx, tenant); err != nil {
			return err
		}
		return svc.IncrementUsage(ctx, tenant, 1)
	})
}
