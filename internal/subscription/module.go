package subscription

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/subscription/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/subscription/repository/sqlc"
)

// Module wires subscription services.
var Module = fx.Options(
	fx.Provide(reposqlc.NewRepository),
	fx.Provide(domain.NewService),
	fx.Provide(RegisterService),
	ModuleGRPC,
	ModuleHTTP,
)
