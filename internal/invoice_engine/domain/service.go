package domain

import (
	"context"
	"strconv"
	"time"

	invoice "github.com/smallbiznis/corebilling/internal/invoice/domain"
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

func (s *Service) GenerateInvoice(ctx context.Context, req *invoiceenginev1.GenerateInvoiceRequest) (*invoiceenginev1.GenerateInvoiceResponse, error) {
	if req == nil || req.GetTenantId() == "" || req.GetSubscriptionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and subscription_id required")
	}

	now := time.Now().UTC()
	invoiceID := strconv.FormatInt(time.Now().UnixNano(), 10)
	start := normalizeTimestamp(req.GetPeriodStart(), now)
	end := normalizeTimestamp(req.GetPeriodEnd(), now)

	inv := invoice.Invoice{
		ID:                 invoiceID,
		TenantID:           req.GetTenantId(),
		BillingPeriodStart: start,
		BillingPeriodEnd:   end,
		TotalCents:         0,
		Status:             "open",
		CreatedAt:          now,
	}

	if err := s.invoiceRepo.Create(ctx, inv); err != nil {
		s.logger.Error("failed to create invoice", zap.Error(err))
		return nil, err
	}

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
		return nil, err
	}

	return &invoiceenginev1.GenerateInvoiceResponse{InvoiceId: invoiceID}, nil
}

func normalizeTimestamp(ts *timestamppb.Timestamp, fallback time.Time) time.Time {
	if ts != nil {
		return ts.AsTime()
	}
	return fallback
}
