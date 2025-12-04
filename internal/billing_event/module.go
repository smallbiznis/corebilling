package billing_event

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/billing_event/domain"
)

// Module wires billing event service into the Fx graph.
var Module = fx.Options(
	fx.Provide(domain.NewService),
	ModuleGRPC,
)

// ModuleGRPC registers the gRPC handler so shared server includes billing event endpoints.
var ModuleGRPC = fx.Invoke(RegisterGRPC)
