package pricing

import (
	"github.com/smallbiznis/corebilling/internal/pricing/domain"
	pricingv1 "github.com/smallbiznis/go-genproto/smallbiznis/pricing/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// ModuleGRPC registers the pricing service with the shared gRPC server.
var ModuleGRPC = fx.Invoke(RegisterGRPC)

// RegisterGRPC attaches the pricing service implementation.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	pricingv1.RegisterPricingServiceServer(server, svc)
}
