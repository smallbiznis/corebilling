package providers

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/events/handler/invoice"
	"github.com/smallbiznis/corebilling/internal/events/handler/subscription"
	"github.com/smallbiznis/corebilling/internal/events/handler/usage"
)

// Module supplies all event handler implementations.
var Module = fx.Options(
	fx.Provide(
		subscription.NewSubscriptionCreatedHandler,
		subscription.NewSubscriptionUpgradedHandler,
		usage.NewUsageReportedHandler,
		usage.NewUsageRatedHandler,
		invoice.NewInvoiceGeneratedHandler,
	),
)
