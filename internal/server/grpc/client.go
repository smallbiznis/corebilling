package grpcserver

import (
	"context"

	"go.uber.org/fx"
)

var ModuleClient = fx.Options(
	fx.Invoke(RegisterHooks),
)

func RegisterClient(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}

var ModuleClientStream = fx.Options(
	fx.Invoke(RegisterClientStream),
)

func RegisterClientStream(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
