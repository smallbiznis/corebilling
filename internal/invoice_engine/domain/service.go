package domain

import (
	"context"
	"fmt"
	"strconv"
	"time"

	invoice "github.com/smallbiznis/corebilling/internal/invoice/domain"
	invoicev1 "github.com/smallbiznis/go-genproto/smallbiznis/invoice/v1"
	invoiceenginev1 "github.com/smallbiznis/go-genproto/smallbiznis/invoice_engine/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service implements the invoice engine API.
type Service struct {
	invoiceenginev1.UnimplementedInvoiceEngineServiceServer
	runRepo     Repository
	invoiceRepo invoice.Repository
	logger      *zap.Logger
}

// NewService constructs the invoice engine service.
func NewService(runRepo Repository, invoiceRepo invoice.Repository, logger *zap.Logger) *Service {
	return &Service{
		runRepo:     runRepo,
		invoiceRepo: invoiceRepo,
		logger:      logger.Named("invoice_engine.service"),
	}
}

func (s *Service) GenerateInvoice(
	ctx context.Context,
	req *invoiceenginev1.GenerateInvoiceRequest,
) (*invoiceenginev1.GenerateInvoiceResponse, error) {
	if req == nil || req.GetTenantId() == "" || req.GetSubscriptionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and subscription_id required")
	}

	now := time.Now().UTC()
	invoiceID := strconv.FormatInt(time.Now().UnixNano(), 10)
	start := normalizeTimestamp(req.GetPeriodStart(), now.AddDate(0, -1, 0))
	end := normalizeTimestamp(req.GetPeriodEnd(), now)

	// // 1. Fetch Price Details
	// priceID := req.GetPriceId()
	// if priceID == "" {
	// 	// Get from subscription
	// 	return nil, status.Error(codes.InvalidArgument, "price_id required")
	// }

	// price, err := s.pricingRepo.GetPrice(ctx, parseID(priceID))
	// if err != nil {
	// 	return nil, status.Errorf(codes.NotFound, "price not found: %v", err)
	// }

	// // 2. Calculate Base Amount (recurring)
	// subtotal := price.UnitAmountCents

	// // 3. Aggregate Usage (if usage-based)
	// usageRecords, _, err := s.usageRepo.List(ctx, usage.ListUsageFilter{
	// 	TenantID:       req.GetTenantId(),
	// 	SubscriptionID: req.GetSubscriptionId(),
	// 	From:           start,
	// 	To:             end,
	// })
	// if err != nil {
	// 	s.logger.Error("failed to fetch usage", zap.Error(err))
	// }

	// // Calculate usage charges
	// var usageCharges int64
	// for _, record := range usageRecords {
	// 	// Simple calculation: value * unit_amount
	// 	// TODO: Apply price tiers if applicable
	// 	usageCharges += int64(record.Value * float64(price.UnitAmountCents) / 1000)
	// }

	// // 4. Apply Price Tiers (if configured)
	// if len(usageRecords) > 0 {
	// 	tiers, err := s.pricingRepo.ListPriceTiersByPriceIDs(ctx, []int64{price.ID})
	// 	if err == nil && len(tiers) > 0 {
	// 		usageCharges = s.calculateTieredCharges(usageRecords, tiers)
	// 	}
	// }

	// 5. Calculate Tax (simplified 10%)
	// subtotalWithUsage := subtotal + usageCharges
	// taxCents := int64(float64(subtotalWithUsage) * 0.10)
	// totalCents := subtotalWithUsage + taxCents

	// 6. Generate Invoice Number
	invoiceNumber := s.generateInvoiceNumber(req.GetTenantId(), now)

	// 7. Create Invoice
	inv := invoice.Invoice{
		ID:             invoiceID,
		TenantID:       req.GetTenantId(),
		CustomerID:     req.GetCustomerId(),
		SubscriptionID: req.GetSubscriptionId(),
		Status:         int32(invoicev1.InvoiceStatus_INVOICE_STATUS_OPEN),
		CurrencyCode:   "USD",
		TotalCents:     0,
		SubtotalCents:  100,
		TaxCents:       0,
		InvoiceNumber:  invoiceNumber,
		IssuedAt:       &start,
		DueAt:          &end,
		// Metadata: map[string]interface{}{
		// 	"price_id":       priceID,
		// 	"base_amount":    subtotal,
		// 	"usage_charges":  usageCharges,
		// 	"usage_count":    len(usageRecords),
		// 	"billing_period": "monthly",
		// },
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.invoiceRepo.Create(ctx, inv); err != nil {
		s.logger.Error("failed to create invoice", zap.Error(err))
		return nil, err
	}

	// 8. Record Engine Run
	run := Run{
		ID:             strconv.FormatInt(time.Now().UnixNano(), 10),
		TenantID:       req.GetTenantId(),
		CustomerID:     req.GetCustomerId(),
		SubscriptionID: req.GetSubscriptionId(),
		InvoiceID:      invoiceID,
		PeriodStart:    start,
		PeriodEnd:      end,
		CreatedAt:      now,
	}

	if err := s.runRepo.Create(ctx, run); err != nil {
		s.logger.Error("failed to record invoice engine run", zap.Error(err))
	}

	// s.logger.Info("invoice generated",
	// 	zap.String("invoice_id", invoiceID),
	// 	zap.String("subscription_id", req.GetSubscriptionId()),
	// 	zap.Int64("total_cents", totalCents),
	// )

	return &invoiceenginev1.GenerateInvoiceResponse{InvoiceId: invoiceID}, nil
}

// calculateTieredCharges applies tiered pricing to usage
// func (s *Service) calculateTieredCharges(
// 	records []usage.UsageRecord,
// 	tiers []pricing.PriceTier,
// ) int64 {
// 	// Sum all usage
// 	var totalUsage float64
// 	for _, r := range records {
// 		totalUsage += r.Value
// 	}

// 	// Apply tiers
// 	var charges int64
// 	remaining := totalUsage

// 	for _, tier := range tiers {
// 		if remaining <= 0 {
// 			break
// 		}

// 		// Calculate units in this tier
// 		tierSize := tier.EndQuantity - tier.StartQuantity
// 		if tierSize <= 0 {
// 			tierSize = remaining // Unlimited tier
// 		}

// 		unitsInTier := min(remaining, tierSize)
// 		charges += int64(unitsInTier * float64(tier.UnitAmountCents))
// 		remaining -= unitsInTier
// 	}

// 	return charges
// }

// generateInvoiceNumber creates unique invoice number
func (s *Service) generateInvoiceNumber(tenantID string, now time.Time) string {
	return fmt.Sprintf("INV-%s-%s-%d",
		tenantID[len(tenantID)-6:], // Last 6 digits of tenant
		now.Format("200601"),       // YYYYMM
		now.Unix()%10000,           // Sequential
	)
}

func normalizeTimestamp(ts *timestamppb.Timestamp, fallback time.Time) time.Time {
	if ts != nil {
		return ts.AsTime()
	}
	return fallback
}
