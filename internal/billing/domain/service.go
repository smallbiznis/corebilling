package domain

import (
	"context"
	"strconv"
	"time"

	billingv1 "github.com/smallbiznis/go-genproto/smallbiznis/billing/v1"
	"go.uber.org/zap"
)

// Service implements the BillingServiceServer API.
type Service struct {
	billingv1.UnimplementedBillingServiceServer
	repo   Repository
	logger *zap.Logger
}

// NewService constructs a billing gRPC service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("billing.service")}
}

// TriggerBilling records a billing run and returns acceptance.
func (s *Service) TriggerBilling(ctx context.Context, req *billingv1.TriggerBillingRequest) (*billingv1.TriggerBillingResponse, error) {
	run := BillingRun{
		ID:             time.Now().UnixNano(),
		TenantID:       parseID(req.GetTenantId()),
		SubscriptionID: parseID(req.GetSubscriptionId()),
		PeriodStart:    req.GetPeriodStart().AsTime(),
		PeriodEnd:      req.GetPeriodEnd().AsTime(),
		Status:         1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, run); err != nil {
		s.logger.Error("failed to persist billing run", zap.Error(err))
		return nil, err
	}

	return &billingv1.TriggerBillingResponse{Accepted: true}, nil
}

func parseID(raw string) int64 {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return id
}
