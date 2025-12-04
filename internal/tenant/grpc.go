package tenant

import (
	"github.com/smallbiznis/corebilling/internal/tenant/domain"
	tenantv1 "github.com/smallbiznis/go-genproto/smallbiznis/tenant/v1"
	"google.golang.org/grpc"
)

// RegisterGRPC registers the tenant service.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	tenantv1.RegisterTenantServiceServer(server, svc)
}
