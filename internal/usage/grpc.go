package usage

import (
	"context"

	"github.com/smallbiznis/corebilling/internal/usage/domain"
	usagev1 "github.com/smallbiznis/go-genproto/smallbiznis/usage/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ModuleGRPC registers the usage service with the shared gRPC server.
var ModuleGRPC = fx.Invoke(RegisterGRPC)

// RegisterGRPC attaches the usage handler.
func RegisterGRPC(server *grpc.Server, svc *domain.Service) {
	usagev1.RegisterUsageServiceServer(server, &grpcService{svc: svc})
}

type grpcService struct {
	usagev1.UnimplementedUsageServiceServer
	svc *domain.Service
}

func (g *grpcService) IngestUsage(ctx context.Context, req *usagev1.IngestUsageRequest) (*usagev1.IngestUsageResponse, error) {
	record := req.GetRecord()
	if record == nil {
		return &usagev1.IngestUsageResponse{}, nil
	}

	model := domain.UsageRecord{
		ID:             record.GetId(),
		SubscriptionID: record.GetSubscriptionId(),
		Metric:         record.GetMeterCode(),
		Quantity:       int64(record.GetValue()),
	}

	if ts := record.GetRecordedAt(); ts != nil {
		model.RecordedAt = ts.AsTime()
	}

	if err := g.svc.Create(ctx, model); err != nil {
		return nil, err
	}

	return &usagev1.IngestUsageResponse{}, nil
}

func (g *grpcService) ListUsage(ctx context.Context, req *usagev1.ListUsageRequest) (*usagev1.ListUsageResponse, error) {
	items, err := g.svc.ListBySubscription(ctx, req.GetSubscriptionId())
	if err != nil {
		return nil, err
	}

	resp := &usagev1.ListUsageResponse{}
	for _, item := range items {
		resp.UsageRecords = append(resp.UsageRecords, &usagev1.UsageRecord{
			Id:             item.ID,
			SubscriptionId: item.SubscriptionID,
			MeterCode:      item.Metric,
			Value:          float64(item.Quantity),
			RecordedAt:     timestamppb.New(item.RecordedAt),
		})
	}
	return resp, nil
}
