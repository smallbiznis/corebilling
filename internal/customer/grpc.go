package customer

import (
	"github.com/smallbiznis/corebilling/internal/customer/domain"
	customerv1 "github.com/smallbiznis/go-genproto/smallbiznis/customer/v1"
	"google.golang.org/grpc"
)

// RegisterGRPC registers the customer service with the gRPC server.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	customerv1.RegisterCustomerServiceServer(server, svc)
}
