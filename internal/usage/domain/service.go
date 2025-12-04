package domain

import (
	"context"

	"go.uber.org/zap"
)

// Service handles usage operations.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService constructs Service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("usage.service")}
}

// Create stores usage data.
func (s *Service) Create(ctx context.Context, usage UsageRecord) error {
	if err := s.repo.Create(ctx, usage); err != nil {
		s.logger.Error("create usage", zap.Error(err))
		return err
	}
	s.logger.Info("usage recorded", zap.String("id", usage.ID))
	return nil
}

// List returns usage records matching the filter.
func (s *Service) List(ctx context.Context, filter ListUsageFilter) ([]UsageRecord, bool, error) {
	return s.repo.List(ctx, filter)
}
