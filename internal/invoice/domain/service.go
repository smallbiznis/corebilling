package domain

import (
	"context"

	"go.uber.org/zap"
)

// Service exposes invoice operations.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService constructs Service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("invoice.service")}
}

// Create stores an invoice.
func (s *Service) Create(ctx context.Context, invoice Invoice) error {
	if err := s.repo.Create(ctx, invoice); err != nil {
		s.logger.Error("create invoice", zap.Error(err))
		return err
	}
	s.logger.Info("invoice created", zap.String("id", invoice.ID))
	return nil
}

// Get retrieves invoice.
func (s *Service) Get(ctx context.Context, id string) (Invoice, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns invoices matching the filter.
func (s *Service) List(ctx context.Context, filter ListInvoicesFilter) ([]Invoice, bool, error) {
	return s.repo.List(ctx, filter)
}
