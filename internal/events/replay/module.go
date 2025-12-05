package replay

import (
	"go.uber.org/fx"

	"github.com/smallbiznis/corebilling/internal/events/replay/repository"
)

var Module = fx.Module(
	"event_replay",
	fx.Provide(
		repository.NewRepository,
		NewService,
	),
	fx.Invoke(RegisterRoutes),
)
