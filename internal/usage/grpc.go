package usage

import (
	"context"
	"strconv"
	"time"

	"github.com/oklog/ulid/v2"
	"github.com/smallbiznis/corebilling/internal/usage/domain"
	usagev1 "github.com/smallbiznis/go-genproto/smallbiznis/usage/v1"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
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

const (
	defaultUsagePageSize = 50
	maxUsagePageSize     = 500
)

func (g *grpcService) IngestUsage(ctx context.Context, req *usagev1.IngestUsageRequest) (*usagev1.IngestUsageResponse, error) {
	record := req.GetRecord()
	if record == nil {
		return nil, status.Error(codes.InvalidArgument, "record required")
	}
	if record.GetTenantId() == "" || record.GetCustomerId() == "" || record.GetSubscriptionId() == "" || record.GetMeterCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id, customer_id, subscription_id, and meter_code are required")
	}

	now := time.Now().UTC()
	usageID := record.GetId()
	if usageID == "" {
		usageID = ulid.Make().String()
	}
	recordedAt := now
	if ts := record.GetRecordedAt(); ts != nil {
		recordedAt = ts.AsTime()
	}

	usage := domain.UsageRecord{
		ID:             usageID,
		TenantID:       record.GetTenantId(),
		CustomerID:     record.GetCustomerId(),
		SubscriptionID: record.GetSubscriptionId(),
		MeterCode:      record.GetMeterCode(),
		Value:          record.GetValue(),
		RecordedAt:     recordedAt,
		IdempotencyKey: record.GetIdempotencyKey(),
		Metadata:       structToMap(record.GetMetadata()),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := g.svc.Create(ctx, usage); err != nil {
		return nil, err
	}

	return &usagev1.IngestUsageResponse{Id: usage.ID}, nil
}

func (g *grpcService) ListUsage(ctx context.Context, req *usagev1.ListUsageRequest) (*usagev1.ListUsageResponse, error) {
	if req == nil || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id required")
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = defaultUsagePageSize
	}
	if pageSize > maxUsagePageSize {
		pageSize = maxUsagePageSize
	}

	filter := domain.ListUsageFilter{
		TenantID:       req.GetTenantId(),
		SubscriptionID: req.GetSubscriptionId(),
		CustomerID:     req.GetCustomerId(),
		MeterCode:      req.GetMeterCode(),
		Limit:          pageSize,
		Offset:         parsePageToken(req.GetPageToken()),
	}
	if from := req.GetFrom(); from != nil {
		filter.From = from.AsTime()
	}
	if to := req.GetTo(); to != nil {
		filter.To = to.AsTime()
	}

	records, hasMore, err := g.svc.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	resp := &usagev1.ListUsageResponse{}
	for _, record := range records {
		resp.Records = append(resp.Records, &usagev1.UsageRecord{
			Id:             record.ID,
			TenantId:       record.TenantID,
			CustomerId:     record.CustomerID,
			SubscriptionId: record.SubscriptionID,
			MeterCode:      record.MeterCode,
			Value:          record.Value,
			RecordedAt:     timestamppb.New(record.RecordedAt),
			IdempotencyKey: record.IdempotencyKey,
			Metadata:       mapToStruct(record.Metadata),
		})
	}
	if hasMore {
		resp.NextPageToken = strconv.Itoa(filter.Offset + len(records))
	}
	return resp, nil
}

func parsePageToken(token string) int {
	if token == "" {
		return 0
	}
	val, err := strconv.Atoi(token)
	if err != nil || val < 0 {
		return 0
	}
	return val
}

func structToMap(value *structpb.Struct) map[string]interface{} {
	if value == nil {
		return nil
	}
	return value.AsMap()
}

func mapToStruct(value map[string]interface{}) *structpb.Struct {
	if len(value) == 0 {
		return nil
	}
	s, err := structpb.NewStruct(value)
	if err != nil {
		return nil
	}
	return s
}
