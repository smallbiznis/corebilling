package webhook

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/webhook/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/webhook/repository/sqlc"
)

var Module = fx.Options(
	fx.Provide(reposqlc.NewRepository),
	fx.Provide(domain.NewService),
	ModuleGRPC,
)

var ModuleGRPC = fx.Invoke(RegisterGRPC)
