package domain

import (
	"context"

	"go.uber.org/zap"
)

// Service exposes billing operations.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService creates a billing service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("billing.service")}
}

// Create records a billing entry.
func (s *Service) Create(ctx context.Context, record BillingRecord) error {
	if err := s.repo.Create(ctx, record); err != nil {
		s.logger.Error("failed to create billing record", zap.Error(err))
		return err
	}
	s.logger.Info("billing record created", zap.String("id", record.ID))
	return nil
}

// Get retrieves a billing record by ID.
func (s *Service) Get(ctx context.Context, id string) (BillingRecord, error) {
	return s.repo.GetByID(ctx, id)
}

// ListByTenant returns billing records for a tenant.
func (s *Service) ListByTenant(ctx context.Context, tenantID string) ([]BillingRecord, error) {
	return s.repo.ListByTenant(ctx, tenantID)
}
