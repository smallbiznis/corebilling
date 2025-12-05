package quota

import (
	"context"

	"github.com/smallbiznis/corebilling/internal/events/outbox"
	"github.com/smallbiznis/corebilling/internal/quota/repository"
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

// Service enforces tenant quota checks and records usage.
type Service struct {
	repo   repository.Repository
	outbox outbox.OutboxRepository
	logger *zap.Logger
}

// NewService constructs a quota service.
func NewService(repo repository.Repository, outboxRepo outbox.OutboxRepository, logger *zap.Logger) *Service {
	return &Service{repo: repo, outbox: outboxRepo, logger: logger.Named("quota.service")}
}

// CheckQuotaForEvent validates usage against limits and emits notifications.
func (s *Service) CheckQuotaForEvent(ctx context.Context, tenantID string) error {
	limit, err := s.repo.GetQuotaLimit(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to load quota limit", zap.Error(err), zap.String("tenant_id", tenantID))
		return err
	}

	usage, err := s.repo.GetQuotaUsage(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to load quota usage", zap.Error(err), zap.String("tenant_id", tenantID))
		return err
	}

	warnLimitEvents := int64(float64(limit.MaxEventsPerDay) * limit.SoftWarningThreshold)
	warnLimitUnits := int64(float64(limit.MaxUsageUnits) * limit.SoftWarningThreshold)

	if usage.EventsToday >= warnLimitEvents || usage.UsageUnits >= warnLimitUnits {
		_ = s.emitEvent(ctx, tenantID, "tenant.quota.warning", usage, limit)
	}

	if usage.EventsToday >= limit.MaxEventsPerDay || usage.UsageUnits >= limit.MaxUsageUnits {
		_ = s.emitEvent(ctx, tenantID, "tenant.quota.exceeded", usage, limit)
		return ErrQuotaExceeded
	}

	return nil
}

// IncrementUsage updates usage counters for a tenant with the provided weight.
func (s *Service) IncrementUsage(ctx context.Context, tenantID string, weight int64) error {
	current, err := s.repo.GetQuotaUsage(ctx, tenantID)
	if err != nil {
		s.logger.Error("failed to load quota usage for increment", zap.Error(err), zap.String("tenant_id", tenantID))
		return err
	}

	params := UpsertQuotaUsageParams{
		TenantID:    tenantID,
		EventsToday: current.EventsToday + 1,
		UsageUnits:  current.UsageUnits + weight,
	}

	if err := s.repo.UpsertQuotaUsage(ctx, params); err != nil {
		s.logger.Error("failed to upsert quota usage", zap.Error(err), zap.String("tenant_id", tenantID))
		return err
	}
	return nil
}

func (s *Service) emitEvent(ctx context.Context, tenantID, subject string, usage QuotaUsage, limit QuotaLimit) error {
	evt := &eventv1.Event{
		Subject:  subject,
		TenantId: tenantID,
		Metadata: mapToStruct(map[string]interface{}{
			"events_today":           usage.EventsToday,
			"usage_units":            usage.UsageUnits,
			"max_events_per_day":     limit.MaxEventsPerDay,
			"max_usage_units":        limit.MaxUsageUnits,
			"soft_warning_threshold": limit.SoftWarningThreshold,
		}),
	}

	if err := s.outbox.InsertOutboxEvent(ctx, &outbox.OutboxEvent{Subject: subject, TenantID: tenantID, Event: evt}); err != nil {
		s.logger.Error("failed to emit quota event", zap.Error(err), zap.String("subject", subject), zap.String("tenant_id", tenantID))
		return err
	}
	return nil
}

func mapToStruct(values map[string]interface{}) *structpb.Struct {
	if values == nil {
		return nil
	}
	fields := make(map[string]*structpb.Value)
	for k, v := range values {
		switch val := v.(type) {
		case string:
			fields[k] = structpb.NewStringValue(val)
		case int64:
			fields[k] = structpb.NewNumberValue(float64(val))
		case float64:
			fields[k] = structpb.NewNumberValue(val)
		case int:
			fields[k] = structpb.NewNumberValue(float64(val))
		}
	}
	return &structpb.Struct{Fields: fields}
}
