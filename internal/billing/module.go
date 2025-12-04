package billing

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/billing/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/billing/repository/sqlc"
)

// Module wires billing domain services.
var Module = fx.Options(
	fx.Provide(reposqlc.NewRepository),
	fx.Provide(domain.NewService),
	ModuleGRPC,
)
