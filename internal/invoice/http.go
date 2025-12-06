package invoice

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	invoicev1 "github.com/smallbiznis/go-genproto/smallbiznis/invoice/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

var ModuleHTTP = fx.Invoke(RegisterHTTP)

func RegisterHTTP(lc fx.Lifecycle, s *grpc.Server, mux *runtime.ServeMux, svc *grpcService) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := invoicev1.RegisterInvoiceServiceHandlerServer(ctx, mux, svc); err != nil {
				return err
			}
			return nil
		},
	})
}
