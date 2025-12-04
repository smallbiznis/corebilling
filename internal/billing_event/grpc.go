package billing_event

import (
	"context"

	"github.com/smallbiznis/corebilling/internal/billing_event/domain"
	billingeventv1 "github.com/smallbiznis/go-genproto/smallbiznis/billing_event/v1"
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegisterGRPC attaches the service to the shared gRPC server.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	billingeventv1.RegisterBillingEventServiceServer(server, &grpcService{svc: svc})
}

type grpcService struct {
	billingeventv1.UnimplementedBillingEventServiceServer
	svc *domain.Service
}

func (g *grpcService) EmitEvent(ctx context.Context, evt *eventv1.Event) (*billingeventv1.BillingEvent, error) {
	if evt == nil || evt.GetSubject() == "" {
		return nil, status.Error(codes.InvalidArgument, "event subject is required")
	}
	return g.svc.EmitEvent(ctx, evt)
}

func (g *grpcService) DeliverWebhook(ctx context.Context, req *billingeventv1.DeliverWebhookRequest) (*billingeventv1.DeliverWebhookResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request required")
	}
	return g.svc.DeliverWebhook(ctx, req)
}
