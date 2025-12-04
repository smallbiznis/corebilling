package meter

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/meter/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/meter/repository/sqlc"
)

var Module = fx.Options(
	fx.Provide(reposqlc.NewRepository),
	fx.Provide(domain.NewService),
	ModuleGRPC,
)

var ModuleGRPC = fx.Invoke(RegisterGRPC)
