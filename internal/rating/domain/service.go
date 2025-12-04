package domain

import (
	"context"

	"go.uber.org/zap"
)

// Service handles rating operations.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService constructs rating service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("rating.service")}
}

// Create stores a rating result.
func (s *Service) Create(ctx context.Context, rating RatingResult) error {
	if err := s.repo.Create(ctx, rating); err != nil {
		s.logger.Error("create rating", zap.Error(err))
		return err
	}
	s.logger.Info("rating created", zap.String("id", rating.ID))
	return nil
}

// GetByUsage returns rating results for a usage record.
func (s *Service) GetByUsage(ctx context.Context, usageID string) ([]RatingResult, error) {
	return s.repo.GetByUsage(ctx, usageID)
}
