package webhook

import (
	"github.com/smallbiznis/corebilling/internal/webhook/domain"
	webhookv1 "github.com/smallbiznis/go-genproto/smallbiznis/webhook/v1"
	"google.golang.org/grpc"
)

// RegisterGRPC registers the webhook service.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	webhookv1.RegisterWebhookServiceServer(server, svc)
}
