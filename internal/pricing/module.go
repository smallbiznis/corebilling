package pricing

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/pricing/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/pricing/repository/sqlc"
)

// Module wires pricing services.
var Module = fx.Options(
        fx.Provide(reposqlc.NewRepository),
        fx.Provide(domain.NewService),
        ModuleGRPC,
)
