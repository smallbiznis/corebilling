package tenant

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/smallbiznis/corebilling/internal/tenant/domain"
	tenantv1 "github.com/smallbiznis/go-genproto/smallbiznis/tenant/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func RegisterHTTP(lc fx.Lifecycle, s *grpc.Server, mux *runtime.ServeMux, svc *domain.Service) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := tenantv1.RegisterTenantServiceHandlerServer(ctx, mux, svc); err != nil {
				return err
			}
			return nil
		},
	})
}
