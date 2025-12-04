package audit

import (
	"github.com/smallbiznis/corebilling/internal/audit/domain"
	auditv1 "github.com/smallbiznis/go-genproto/smallbiznis/audit/v1"
	"google.golang.org/grpc"
)

// RegisterGRPC registers the audit service with the gRPC server.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	auditv1.RegisterAuditServiceServer(server, svc)
}
