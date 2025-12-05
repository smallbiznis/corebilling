package billingcycle

import (
	"context"
	"time"

	"github.com/smallbiznis/corebilling/internal/billingcycle/repository"
	"github.com/smallbiznis/corebilling/internal/events/outbox"
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// InvoiceGenerator triggers downstream invoice creation.
type InvoiceGenerator interface {
	GenerateForTenant(ctx context.Context, tenantID string) error
}

// Service handles billing cycle transitions.
type Service struct {
	repo      repository.Repository
	generator InvoiceGenerator
	outbox    outbox.OutboxRepository
	logger    *zap.Logger
}

// NewService constructs the billing cycle service.
func NewService(repo repository.Repository, generator InvoiceGenerator, outboxRepo outbox.OutboxRepository, logger *zap.Logger) *Service {
	return &Service{repo: repo, generator: generator, outbox: outboxRepo, logger: logger.Named("billingcycle.service")}
}

// CloseBillingCycle finalizes the current window and advances to the next.
func (s *Service) CloseBillingCycle(ctx context.Context, tenantID string) error {
	cycle, err := s.repo.GetCycleForTenant(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to load billing cycle", zap.Error(err), zap.String("tenant_id", tenantID))
		return err
	}

	if err := s.emitClosedEvent(ctx, cycle); err != nil {
		return err
	}

	if s.generator != nil {
		if err := s.generator.GenerateForTenant(ctx, tenantID); err != nil {
			s.logger.Error("invoice generation failed", zap.Error(err), zap.String("tenant_id", tenantID))
		}
	}

	newStart := cycle.PeriodEnd
	newEnd := newStart.AddDate(0, 1, 0)

	if err := s.repo.UpdateBillingCycle(ctx, UpdateBillingCycleParams{
		TenantID:    tenantID,
		PeriodStart: newStart,
		PeriodEnd:   newEnd,
	}); err != nil {
		s.logger.Error("failed to update billing cycle", zap.Error(err), zap.String("tenant_id", tenantID))
		return err
	}

	return nil
}

func (s *Service) emitClosedEvent(ctx context.Context, cycle BillingCycle) error {
	evt := &eventv1.Event{
		Subject:  "billing_cycle.closed",
		TenantId: cycle.TenantID,
		Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{
			"period_start": structpb.NewStringValue(cycle.PeriodStart.Format(time.RFC3339)),
			"period_end":   structpb.NewStringValue(cycle.PeriodEnd.Format(time.RFC3339)),
		}},
	}

	if err := s.outbox.InsertOutboxEvent(ctx, &outbox.OutboxEvent{Subject: evt.Subject, TenantID: cycle.TenantID, Event: evt}); err != nil {
		s.logger.Error("failed to persist billing cycle event", zap.Error(err), zap.String("tenant_id", cycle.TenantID))
		return err
	}
	return nil
}
