package domain

import (
	"context"
	"strconv"
	"time"

	pricingv1 "github.com/smallbiznis/go-genproto/smallbiznis/pricing/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service implements PricingServiceServer.
type Service struct {
	pricingv1.UnimplementedPricingServiceServer
	repo   Repository
	logger *zap.Logger
}

// NewService constructs service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("pricing.service")}
}

// CreateProduct stores a product.
func (s *Service) CreateProduct(ctx context.Context, req *pricingv1.CreateProductRequest) (*pricingv1.Product, error) {
	now := time.Now()
	id := time.Now().UnixNano()
	p := req.GetProduct()
	model := Product{
		ID:          id,
		TenantID:    parseID(p.GetTenantId()),
		Name:        p.GetName(),
		Code:        p.GetCode(),
		Description: p.GetDescription(),
		Active:      p.GetActive(),
		Metadata:    structToMap(p.GetMetadata()),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.repo.CreateProduct(ctx, model); err != nil {
		return nil, err
	}
	return s.toProductProto(model), nil
}

// GetProduct fetches a product.
func (s *Service) GetProduct(ctx context.Context, req *pricingv1.GetProductRequest) (*pricingv1.Product, error) {
	prod, err := s.repo.GetProduct(ctx, parseID(req.GetId()))
	if err != nil {
		return nil, err
	}
	return s.toProductProto(prod), nil
}

// ListProducts lists products by tenant.
func (s *Service) ListProducts(ctx context.Context, req *pricingv1.ListProductsRequest) (*pricingv1.ListProductsResponse, error) {
	items, err := s.repo.ListProducts(ctx, parseID(req.GetTenantId()))
	if err != nil {
		return nil, err
	}
	resp := &pricingv1.ListProductsResponse{}
	for _, p := range items {
		resp.Products = append(resp.Products, s.toProductProto(p))
	}
	return resp, nil
}

// CreatePrice persists a price.
func (s *Service) CreatePrice(ctx context.Context, req *pricingv1.CreatePriceRequest) (*pricingv1.Price, error) {
	now := time.Now()
	id := time.Now().UnixNano()
	p := req.GetPrice()
	model := Price{
		ID:                   id,
		TenantID:             parseID(p.GetTenantId()),
		ProductID:            parseID(p.GetProductId()),
		Code:                 p.GetCode(),
		LookupKey:            p.GetLookupKey(),
		PricingModel:         int32(p.GetPricingModel()),
		Currency:             p.GetCurrency(),
		UnitAmountCents:      p.GetUnitAmountCents(),
		BillingInterval:      int32(p.GetBillingInterval()),
		BillingIntervalCount: p.GetBillingIntervalCount(),
		Active:               p.GetActive(),
		Metadata:             structToMap(p.GetMetadata()),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := s.repo.CreatePrice(ctx, model); err != nil {
		return nil, err
	}
	return s.toPriceProto(model), nil
}

// GetPrice loads a price.
func (s *Service) GetPrice(ctx context.Context, req *pricingv1.GetPriceRequest) (*pricingv1.Price, error) {
	price, err := s.repo.GetPrice(ctx, parseID(req.GetId()))
	if err != nil {
		return nil, err
	}
	return s.toPriceProto(price), nil
}

// ListPrices returns prices for a product.
func (s *Service) ListPrices(ctx context.Context, req *pricingv1.ListPricesRequest) (*pricingv1.ListPricesResponse, error) {
	items, err := s.repo.ListPrices(ctx, parseID(req.GetProductId()))
	if err != nil {
		return nil, err
	}
	resp := &pricingv1.ListPricesResponse{}
	for _, p := range items {
		resp.Prices = append(resp.Prices, s.toPriceProto(p))
	}
	return resp, nil
}

func (s *Service) toProductProto(p Product) *pricingv1.Product {
	metadata, _ := structpb.NewStruct(p.Metadata)
	return &pricingv1.Product{
		Id:          strconv.FormatInt(p.ID, 10),
		TenantId:    strconv.FormatInt(p.TenantID, 10),
		Name:        p.Name,
		Code:        p.Code,
		Description: p.Description,
		Active:      p.Active,
		Metadata:    metadata,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

func (s *Service) toPriceProto(p Price) *pricingv1.Price {
	metadata, _ := structpb.NewStruct(p.Metadata)
	return &pricingv1.Price{
		Id:                   strconv.FormatInt(p.ID, 10),
		TenantId:             strconv.FormatInt(p.TenantID, 10),
		ProductId:            strconv.FormatInt(p.ProductID, 10),
		Code:                 p.Code,
		LookupKey:            p.LookupKey,
		PricingModel:         pricingv1.PricingModel(p.PricingModel),
		Currency:             p.Currency,
		UnitAmountCents:      p.UnitAmountCents,
		BillingInterval:      pricingv1.BillingInterval(p.BillingInterval),
		BillingIntervalCount: p.BillingIntervalCount,
		Active:               p.Active,
		Metadata:             metadata,
		CreatedAt:            timestamppb.New(p.CreatedAt),
		UpdatedAt:            timestamppb.New(p.UpdatedAt),
	}
}

func structToMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}
	return s.AsMap()
}

func parseID(raw string) int64 {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
