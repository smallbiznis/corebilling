package rating

import (
	"context"

	"github.com/smallbiznis/corebilling/internal/rating/domain"
	ratingv1 "github.com/smallbiznis/go-genproto/smallbiznis/rating/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ModuleGRPC registers the rating service with the shared gRPC server.
var ModuleGRPC = fx.Invoke(RegisterGRPC)

// RegisterGRPC attaches the rating handler.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	ratingv1.RegisterRatingServiceServer(server, &grpcService{svc: svc})
}

type grpcService struct {
	ratingv1.UnimplementedRatingServiceServer
	svc *domain.Service
}

func (g *grpcService) RateSubscription(ctx context.Context, req *ratingv1.RateSubscriptionRequest) (*ratingv1.RateSubscriptionResponse, error) {
	// Domain logic is not yet implemented; return unimplemented to satisfy the interface.
	return nil, status.Error(codes.Unimplemented, "rating not implemented")
}
