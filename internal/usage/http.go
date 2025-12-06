package usage

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	usagev1 "github.com/smallbiznis/go-genproto/smallbiznis/usage/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var ModuleHTTP = fx.Invoke(RegisterHTTP)

func RegisterHTTP(lc fx.Lifecycle, s *grpc.Server, mux *runtime.ServeMux, svc *grpcService) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := usagev1.RegisterUsageServiceHandlerServer(ctx, mux, svc); err != nil {
				return err
			}
			return nil
		},
	})
}
