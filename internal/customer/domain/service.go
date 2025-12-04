package domain

import (
	"context"
	"strconv"
	"time"

	customerv1 "github.com/smallbiznis/go-genproto/smallbiznis/customer/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const defaultCustomerPageSize = 25

// Service implements the CustomerServiceServer API.
type Service struct {
	customerv1.UnimplementedCustomerServiceServer
	repo   Repository
	logger *zap.Logger
}

// NewService injects dependencies for customer operations.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("customer.service")}
}

func (s *Service) CreateCustomer(ctx context.Context, req *customerv1.CreateCustomerRequest) (*customerv1.Customer, error) {
	payload := req.GetCustomer()
	if payload == nil || payload.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "customer payload and tenant_id are required")
	}

	id := payload.GetId()
	if id == "" {
		id = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	now := time.Now().UTC()
	customer := Customer{
		ID:                id,
		TenantID:          payload.GetTenantId(),
		ExternalReference: payload.GetExternalReference(),
		Email:             payload.GetEmail(),
		Name:              payload.GetName(),
		Phone:             payload.GetPhone(),
		Currency:          payload.GetCurrency(),
		BillingAddress:    structToMap(payload.GetBillingAddress()),
		ShippingAddress:   structToMap(payload.GetShippingAddress()),
		Metadata:          structToMap(payload.GetMetadata()),
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := s.repo.Create(ctx, customer); err != nil {
		s.logger.Error("failed to persist customer", zap.Error(err))
		return nil, err
	}
	return s.toProto(customer), nil
}

func (s *Service) GetCustomer(ctx context.Context, req *customerv1.GetCustomerRequest) (*customerv1.Customer, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "customer id is required")
	}
	customer, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return s.toProto(customer), nil
}

func (s *Service) ListCustomers(ctx context.Context, req *customerv1.ListCustomersRequest) (*customerv1.ListCustomersResponse, error) {
	if req == nil || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	pageSize := int(req.GetPageSize())
	if pageSize <= 0 {
		pageSize = defaultCustomerPageSize
	}
	offset := parsePageToken(req.GetPageToken())

	customers, err := s.repo.ListByTenant(ctx, req.GetTenantId(), pageSize, offset)
	if err != nil {
		return nil, err
	}

	resp := &customerv1.ListCustomersResponse{}
	for _, item := range customers {
		resp.Customers = append(resp.Customers, s.toProto(item))
	}
	if len(customers) == pageSize {
		resp.NextPageToken = strconv.Itoa(offset + len(customers))
	}
	return resp, nil
}

func (s *Service) UpdateCustomer(ctx context.Context, req *customerv1.UpdateCustomerRequest) (*customerv1.Customer, error) {
	payload := req.GetCustomer()
	if payload == nil || payload.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "customer payload with id is required")
	}
	existing, err := s.repo.GetByID(ctx, payload.GetId())
	if err != nil {
		return nil, err
	}

	if payload.GetTenantId() != "" {
		existing.TenantID = payload.GetTenantId()
	}
	existing.ExternalReference = payload.GetExternalReference()
	existing.Email = payload.GetEmail()
	existing.Name = payload.GetName()
	existing.Phone = payload.GetPhone()
	existing.Currency = payload.GetCurrency()
	existing.BillingAddress = structToMap(payload.GetBillingAddress())
	existing.ShippingAddress = structToMap(payload.GetShippingAddress())
	existing.Metadata = structToMap(payload.GetMetadata())
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		s.logger.Error("customer update failed", zap.Error(err))
		return nil, err
	}
	return s.toProto(existing), nil
}

func (s *Service) toProto(customer Customer) *customerv1.Customer {
	return &customerv1.Customer{
		Id:                customer.ID,
		TenantId:          customer.TenantID,
		ExternalReference: customer.ExternalReference,
		Email:             customer.Email,
		Name:              customer.Name,
		Phone:             customer.Phone,
		Currency:          customer.Currency,
		BillingAddress:    mapToStruct(customer.BillingAddress),
		ShippingAddress:   mapToStruct(customer.ShippingAddress),
		Metadata:          mapToStruct(customer.Metadata),
		CreatedAt:         timestamppb.New(customer.CreatedAt),
		UpdatedAt:         timestamppb.New(customer.UpdatedAt),
	}
}

func structToMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return nil
	}
	return s.AsMap()
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
