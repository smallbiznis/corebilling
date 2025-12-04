package usage

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/usage/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/usage/repository/sqlc"
)

// Module wires usage services.
var Module = fx.Options(
	fx.Provide(reposqlc.NewRepository),
	fx.Provide(domain.NewService),
	ModuleGRPC,
)
