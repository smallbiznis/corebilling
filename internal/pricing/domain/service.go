package domain

import (
	"context"

	"go.uber.org/zap"
)

// Service handles pricing operations.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService constructs Service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("pricing.service")}
}

// Create adds a pricing plan.
func (s *Service) Create(ctx context.Context, plan PricingPlan) error {
	if err := s.repo.Create(ctx, plan); err != nil {
		s.logger.Error("create pricing plan", zap.Error(err))
		return err
	}
	s.logger.Info("pricing plan created", zap.String("id", plan.ID))
	return nil
}

// Get retrieves a plan.
func (s *Service) Get(ctx context.Context, id string) (PricingPlan, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns all plans.
func (s *Service) List(ctx context.Context) ([]PricingPlan, error) {
	return s.repo.List(ctx)
}
