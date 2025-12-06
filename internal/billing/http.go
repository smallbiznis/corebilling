package billing

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/smallbiznis/corebilling/internal/billing/domain"
	billingv1 "github.com/smallbiznis/go-genproto/smallbiznis/billing/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var ModuleHTTP = fx.Invoke(RegisterHTTP)

func RegisterHTTP(lc fx.Lifecycle, s *grpc.Server, mux *runtime.ServeMux, svc *domain.Service) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := billingv1.RegisterBillingServiceHandlerServer(ctx, mux, svc); err != nil {
				return err
			}
			return nil
		},
	})
}
