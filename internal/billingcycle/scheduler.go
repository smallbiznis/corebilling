package billingcycle

import (
	"context"
	"time"

	"github.com/smallbiznis/corebilling/internal/billingcycle/repository"
	"go.uber.org/zap"
)

// Scheduler runs periodic billing cycle closings.
type Scheduler struct {
	repo   repository.Repository
	svc    *Service
	logger *zap.Logger
}

// NewScheduler constructs a scheduler.
func NewScheduler(repo repository.Repository, svc *Service, logger *zap.Logger) *Scheduler {
	return &Scheduler{repo: repo, svc: svc, logger: logger.Named("billingcycle.scheduler")}
}

// Run starts the periodic worker.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.process(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Scheduler) process(ctx context.Context) {
	tenants, err := s.repo.ListTenantsDueForCycleClose(ctx)
	if err != nil {
		s.logger.Error("failed to list tenants due for cycle close", zap.Error(err))
		return
	}
	for _, tenant := range tenants {
		if err := s.svc.CloseBillingCycle(ctx, tenant); err != nil {
			s.logger.Error("failed to close billing cycle", zap.Error(err), zap.String("tenant_id", tenant))
		}
	}
}
