package domain

import (
	"context"

	"go.uber.org/zap"
)

// Service handles subscription workflows.
type Service struct {
	repo   Repository
	logger *zap.Logger
}

// NewService constructs Service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("subscription.service")}
}

// Create registers a subscription.
func (s *Service) Create(ctx context.Context, sub Subscription) error {
	if err := s.repo.Create(ctx, sub); err != nil {
		s.logger.Error("create subscription", zap.Error(err))
		return err
	}
	s.logger.Info("subscription created", zap.String("id", sub.ID))
	return nil
}

// Get returns a subscription by id.
func (s *Service) Get(ctx context.Context, id string) (Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

// List returns subscriptions for a tenant.
func (s *Service) List(ctx context.Context, filter ListSubscriptionsFilter) ([]Subscription, bool, error) {
	return s.repo.List(ctx, filter)
}

// Update persists subscription changes.
func (s *Service) Update(ctx context.Context, sub Subscription) error {
	if err := s.repo.Update(ctx, sub); err != nil {
		s.logger.Error("update subscription", zap.Error(err))
		return err
	}
	return nil
}
