package domain

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/smallbiznis/corebilling/internal/headers"
	pricingv1 "github.com/smallbiznis/go-genproto/smallbiznis/pricing/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service implements PricingServiceServer.
type Service struct {
	pricingv1.UnimplementedPricingServiceServer
	repo   Repository
	logger *zap.Logger

	genID *snowflake.Node
}

// NewService constructs service.
func NewService(repo Repository, logger *zap.Logger, genID *snowflake.Node) *Service {
	return &Service{repo: repo, logger: logger.Named("pricing.service"), genID: genID}
}

// CreateProduct stores a product.
func (s *Service) CreateProduct(ctx context.Context, req *pricingv1.CreateProductRequest) (*pricingv1.Product, error) {
	now := time.Now()
	id := s.genID.Generate()
	p := req.GetProduct()

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		s.logger.Info("metadata", zap.Strings("tenantId", md.Get("x-tenant-id")))
	}

	fmt.Printf("TenantID: %v\n", ok)

	tenantID, err := snowflake.ParseString(p.GetTenantId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}

	model := Product{
		ID:          id.Int64(),
		TenantID:    tenantID.Int64(),
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

	md, _ := metadata.FromIncomingContext(ctx)
	h := headers.ExtractMetadata(md)

	tenantID, err := snowflake.ParseString(h.TenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}

	items, err := s.repo.ListProducts(ctx, tenantID.Int64())
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
	id := s.genID.Generate()
	p := req.GetPrice()

	md, _ := metadata.FromIncomingContext(ctx)
	h := headers.ExtractMetadata(md)

	tenantID, err := snowflake.ParseString(h.TenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}

	productID, err := snowflake.ParseString(p.GetProductId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}

	price := Price{
		ID:                   id.Int64(),
		TenantID:             tenantID.Int64(),
		ProductID:            productID.Int64(),
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
	if err := s.repo.CreatePrice(ctx, price); err != nil {
		return nil, err
	}

	for _, t := range req.GetTiers() {
		tier := PriceTier{
			ID:              s.genID.Generate().Int64(),
			PriceID:         price.ID,
			StartQuantity:   t.GetStartQuantity(),
			EndQuantity:     t.GetEndQuantity(),
			UnitAmountCents: t.GetUnitAmountCents(),
			Unit:            t.GetUnit(),
			Metadata:        structToMap(t.GetMetadata()),
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := s.repo.CreatePriceTier(ctx, tier); err != nil {
			return nil, err
		}
	}

	return s.toPriceProto(price), nil
}

// GetPrice loads a price.
func (s *Service) GetPrice(ctx context.Context, req *pricingv1.GetPriceRequest) (*pricingv1.Price, error) {

	md, _ := metadata.FromIncomingContext(ctx)
	h := headers.ExtractMetadata(md)

	tenantID, err := snowflake.ParseString(h.TenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}

	price, err := s.repo.GetPrice(ctx, tenantID.Int64(), parseID(req.GetId()))
	if err != nil {
		return nil, err
	}
	return s.toPriceProto(price), nil
}

// ListPrices returns prices for a product.
func (s *Service) ListPrices(ctx context.Context, req *pricingv1.ListPricesRequest) (*pricingv1.ListPricesResponse, error) {

	md, _ := metadata.FromIncomingContext(ctx)
	h := headers.ExtractMetadata(md)

	tenantID, err := snowflake.ParseString(h.TenantID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid tenant_id")
	}

	items, err := s.repo.ListPrices(ctx, tenantID.Int64(), parseID(req.GetProductId()))
	if err != nil {
		return nil, err
	}
	resp := &pricingv1.ListPricesResponse{}
	var priceIDs []int64
	for _, p := range items {
		resp.Prices = append(resp.Prices, s.toPriceProto(p))
		priceIDs = append(priceIDs, p.ID)
	}

	if len(priceIDs) > 0 {
		tiers, err := s.repo.ListPriceTiersByPriceIDs(ctx, priceIDs)
		if err != nil {
			return nil, err
		}

		for _, tier := range tiers {
			resp.Tiers = append(resp.Tiers, s.toPriceTierProto(tier))
		}
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

func (s *Service) toPriceTierProto(t PriceTier) *pricingv1.PriceTier {
	metadata, _ := structpb.NewStruct(t.Metadata)
	return &pricingv1.PriceTier{
		Id:              strconv.FormatInt(t.ID, 10),
		PriceId:         strconv.FormatInt(t.PriceID, 10),
		StartQuantity:   t.StartQuantity,
		EndQuantity:     t.EndQuantity,
		UnitAmountCents: t.UnitAmountCents,
		Unit:            t.Unit,
		Metadata:        metadata,
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
