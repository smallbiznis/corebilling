package customer

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/smallbiznis/corebilling/internal/customer/domain"
	customerv1 "github.com/smallbiznis/go-genproto/smallbiznis/customer/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func RegisterHTTP(lc fx.Lifecycle, s *grpc.Server, mux *runtime.ServeMux, svc *domain.Service) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := customerv1.RegisterCustomerServiceHandlerServer(ctx, mux, svc); err != nil {
				return err
			}
			return nil
		},
	})
}
