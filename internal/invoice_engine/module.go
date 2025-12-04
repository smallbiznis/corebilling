package invoice_engine

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/invoice_engine/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/invoice_engine/repository/sqlc"
)

var Module = fx.Options(
	fx.Provide(reposqlc.NewRepository),
	fx.Provide(domain.NewService),
	ModuleGRPC,
)

var ModuleGRPC = fx.Invoke(RegisterGRPC)
