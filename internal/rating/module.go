package rating

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/rating/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/rating/repository/sqlc"
)

// Module wires rating services.
var Module = fx.Options(
        fx.Provide(reposqlc.NewRepository),
        fx.Provide(domain.NewService),
        ModuleGRPC,
)
