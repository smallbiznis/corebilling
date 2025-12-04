package billing

import (
	"github.com/smallbiznis/corebilling/internal/billing/domain"
	billingv1 "github.com/smallbiznis/go-genproto/smallbiznis/billing/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// ModuleGRPC registers the billing service implementation with the shared gRPC server.
var ModuleGRPC = fx.Invoke(RegisterGRPC)

// RegisterGRPC attaches the billing service to the server.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	billingv1.RegisterBillingServiceServer(server, svc)
}
