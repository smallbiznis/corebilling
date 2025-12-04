package invoice

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/invoice/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/invoice/repository/sqlc"
)

// Module wires invoice services.
var Module = fx.Options(
        fx.Provide(reposqlc.NewRepository),
        fx.Provide(domain.NewService),
        ModuleGRPC,
)
