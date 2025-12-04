package domain

import (
	"context"
	"strconv"
	"time"

	meterv1 "github.com/smallbiznis/go-genproto/smallbiznis/meter/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const defaultMeterPageSize = 25

// Service implements the meter gRPC API.
type Service struct {
	meterv1.UnimplementedMeterServiceServer
	repo   Repository
	logger *zap.Logger
}

// NewService constructs a meter service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("meter.service")}
}

func (s *Service) CreateMeter(ctx context.Context, req *meterv1.CreateMeterRequest) (*meterv1.CreateMeterResponse, error) {
	payload := req.GetMeter()
	if payload == nil || payload.GetTenantId() == "" || payload.GetCode() == "" {
		return nil, status.Error(codes.InvalidArgument, "meter tenant and code required")
	}
	id := payload.GetId()
	if id == "" {
		id = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	now := time.Now().UTC()
	meter := Meter{
		ID:          id,
		TenantID:    payload.GetTenantId(),
		Code:        payload.GetCode(),
		Aggregation: int32(payload.GetAggregation()),
		Transform:   int32(payload.GetTransform()),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, meter); err != nil {
		s.logger.Error("create meter failed", zap.Error(err))
		return nil, err
	}
	return &meterv1.CreateMeterResponse{Meter: s.toProto(meter)}, nil
}

func (s *Service) GetMeter(ctx context.Context, req *meterv1.GetMeterRequest) (*meterv1.Meter, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "meter id required")
	}

	meter, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return s.toProto(meter), nil
}

func (s *Service) ListMeters(ctx context.Context, req *meterv1.ListMetersRequest) (*meterv1.ListMetersResponse, error) {
	if req == nil || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = defaultMeterPageSize
	}
	offset := parsePageToken(req.GetPageToken())

	meters, err := s.repo.ListByTenant(ctx, req.GetTenantId(), pageSize, offset)
	if err != nil {
		return nil, err
	}

	resp := &meterv1.ListMetersResponse{}
	for _, item := range meters {
		resp.Meters = append(resp.Meters, s.toProto(item))
	}
	if len(meters) == pageSize {
		resp.NextPageToken = strconv.Itoa(offset + len(meters))
	}
	return resp, nil
}

func (s *Service) UpdateMeter(ctx context.Context, req *meterv1.UpdateMeterRequest) (*meterv1.UpdateMeterResponse, error) {
	payload := req.GetMeter()
	if payload == nil || payload.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "meter id required")
	}

	existing, err := s.repo.GetByID(ctx, payload.GetId())
	if err != nil {
		return nil, err
	}

	if payload.GetTenantId() != "" {
		existing.TenantID = payload.GetTenantId()
	}
	if payload.GetCode() != "" {
		existing.Code = payload.GetCode()
	}
	existing.Aggregation = int32(payload.GetAggregation())
	existing.Transform = int32(payload.GetTransform())
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("update meter failed", zap.Error(err))
		return nil, err
	}
	return &meterv1.UpdateMeterResponse{Meter: s.toProto(existing)}, nil
}

func (s *Service) toProto(meter Meter) *meterv1.Meter {
	return &meterv1.Meter{
		Id:          meter.ID,
		TenantId:    meter.TenantID,
		Code:        meter.Code,
		Aggregation: meterv1.MeterAggregation(meter.Aggregation),
		Transform:   meterv1.MeterTransform(meter.Transform),
	}
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
