package app

import (
	"go.uber.org/fx"

	"github.com/bwmarrin/snowflake"
	"github.com/smallbiznis/corebilling/internal/audit"
	"github.com/smallbiznis/corebilling/internal/billing"
	"github.com/smallbiznis/corebilling/internal/billing_event"
	"github.com/smallbiznis/corebilling/internal/billingcycle"
	"github.com/smallbiznis/corebilling/internal/config"
	"github.com/smallbiznis/corebilling/internal/customer"
	"github.com/smallbiznis/corebilling/internal/db"
	"github.com/smallbiznis/corebilling/internal/eventfx"
	"github.com/smallbiznis/corebilling/internal/invoice"
	"github.com/smallbiznis/corebilling/internal/invoice_engine"
	"github.com/smallbiznis/corebilling/internal/ledger"
	"github.com/smallbiznis/corebilling/internal/log"
	"github.com/smallbiznis/corebilling/internal/meter"
	"github.com/smallbiznis/corebilling/internal/pricing"
	"github.com/smallbiznis/corebilling/internal/quota"
	"github.com/smallbiznis/corebilling/internal/rating"
	grpcserver "github.com/smallbiznis/corebilling/internal/server/grpc"
	httpserver "github.com/smallbiznis/corebilling/internal/server/http"
	"github.com/smallbiznis/corebilling/internal/subscription"
	"github.com/smallbiznis/corebilling/internal/telemetry"
	"github.com/smallbiznis/corebilling/internal/tenant"
	"github.com/smallbiznis/corebilling/internal/usage"
	"github.com/smallbiznis/corebilling/internal/webhook"
)

func Cloud() *fx.App {
	return fx.New(
		fx.Provide(config.Load),
		log.Module,
		db.Module,
		snowflakeNode,
		telemetry.Module,
		tenant.Module,
	)
}

func Billing() *fx.App {
	return fx.New(
		fx.Provide(config.Load),
		log.Module,
		db.Module,
		db.MigrationsModule,
		snowflakeNode,
		telemetry.Module,
		eventfx.Module,
		tenant.Module,
		billing.Module,
		billing_event.Module,
		audit.Module,
		customer.Module,
		pricing.Module,
		meter.Module,
		invoice_engine.Module,
		billingcycle.Module,
		quota.Module,
		subscription.Module,
		usage.Module,
		rating.Module,
		ledger.Module,
		invoice.Module,
		grpcserver.Module,
		httpserver.Module,
		webhook.Module,
	)
}

var snowflakeNode = fx.Provide(func() (*snowflake.Node, error) {
	return snowflake.NewNode(1)
})
