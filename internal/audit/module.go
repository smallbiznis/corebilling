package audit

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/audit/domain"
	reposqlc "github.com/smallbiznis/corebilling/internal/audit/repository/sqlc"
)

var Module = fx.Options(
	fx.Provide(reposqlc.NewRepository),
	fx.Provide(domain.NewService),
	ModuleGRPC,
)

var ModuleGRPC = fx.Invoke(RegisterGRPC)
