package domain

import (
	"context"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	tenantv1 "github.com/smallbiznis/go-genproto/smallbiznis/tenant/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultTenantPageSize = 25

// Service implements tenantv1.TenantServiceServer.
type Service struct {
	tenantv1.UnimplementedTenantServiceServer
	repo   Repository
	logger *zap.Logger

	genID *snowflake.Node
}

// NewService constructs a tenant service.
func NewService(repo Repository, logger *zap.Logger, genID *snowflake.Node) *Service {
	return &Service{repo: repo, logger: logger.Named("tenant.service"), genID: genID}
}

func (s *Service) CreateTenant(ctx context.Context, req *tenantv1.CreateTenantRequest) (*tenantv1.Tenant, error) {
	if req == nil || req.GetName() == "" || req.GetSlug() == "" {
		return nil, status.Error(codes.InvalidArgument, "name and slug are required")
	}

	now := time.Now().UTC()
	tenantID := s.genID.Generate()
	tenant := Tenant{
		ID:              tenantID.Int64(),
		Type:            tenantv1.TenantType_TENANT_TYPE_PLATFORM,
		Name:            req.GetName(),
		Slug:            req.GetSlug(),
		Status:          tenantv1.TenantStatus_TENANT_STATUS_ACTIVE,
		DefaultCurrency: req.GetDefaultCurrency(),
		CountryCode:     req.GetCountryCode(),
		Metadata:        structToMap(req.GetMetadata()),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if req.GetParentId() != "" {
		parentID, err := snowflake.ParseString(req.GetParentId())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid parent_id")
		}

		tenant.ParentID = parentID.Int64()
		tenant.Type = tenantv1.TenantType_TENANT_TYPE_CONNECTED_ACCOUNT
	}

	if err := s.repo.Create(ctx, tenant); err != nil {
		s.logger.Error("failed to persist tenant", zap.Error(err))
		return nil, err
	}
	return s.toProto(tenant), nil
}

func (s *Service) GetTenant(ctx context.Context, req *tenantv1.GetTenantRequest) (*tenantv1.Tenant, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	tenant, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return s.toProto(tenant), nil
}

func (s *Service) ListTenants(ctx context.Context, req *tenantv1.ListTenantsRequest) (*tenantv1.ListTenantsResponse, error) {
	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = defaultTenantPageSize
	}
	offset := parsePageToken(req.GetPageToken())
	filter := ListFilter{
		ParentID: req.GetParentId(),
		Limit:    pageSize,
		Offset:   offset,
	}
	items, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	resp := &tenantv1.ListTenantsResponse{}
	for _, item := range items {
		resp.Tenants = append(resp.Tenants, s.toProto(item))
	}
	if len(items) == pageSize {
		resp.NextPageToken = strconv.Itoa(offset + len(items))
	}
	return resp, nil
}

func (s *Service) UpdateTenant(ctx context.Context, req *tenantv1.UpdateTenantRequest) (*tenantv1.Tenant, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	tenant, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	if req.GetName() != "" {
		tenant.Name = req.GetName()
	}
	if req.GetSlug() != "" {
		tenant.Slug = req.GetSlug()
	}
	if req.GetStatus() != tenantv1.TenantStatus_TENANT_STATUS_UNSPECIFIED {
		tenant.Status = req.GetStatus()
	}
	if req.GetDefaultCurrency() != "" {
		tenant.DefaultCurrency = req.GetDefaultCurrency()
	}
	if req.GetCountryCode() != "" {
		tenant.CountryCode = req.GetCountryCode()
	}
	if req.GetMetadata() != nil {
		tenant.Metadata = structToMap(req.GetMetadata())
	}
	tenant.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, tenant); err != nil {
		s.logger.Error("update failed", zap.Error(err))
		return nil, err
	}
	return s.toProto(tenant), nil
}

func (s *Service) toProto(t Tenant) *tenantv1.Tenant {
	metadata, _ := structpb.NewStruct(t.Metadata)
	tenant := &tenantv1.Tenant{
		Id:              strconv.FormatInt(t.ID, 10),
		Type:            t.Type,
		Name:            t.Name,
		Slug:            t.Slug,
		Status:          t.Status,
		DefaultCurrency: t.DefaultCurrency,
		CountryCode:     t.CountryCode,
		Metadata:        metadata,
		CreatedAt:       timestamppb.New(t.CreatedAt),
		UpdatedAt:       timestamppb.New(t.UpdatedAt),
	}

	if t.ParentID > 0 {

	}

	return tenant
}

func structToMap(value *structpb.Struct) map[string]interface{} {
	if value == nil {
		return nil
	}
	return value.AsMap()
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
