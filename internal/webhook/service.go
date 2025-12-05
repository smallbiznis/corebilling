package webhook

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/smallbiznis/corebilling/internal/events"
	"github.com/smallbiznis/corebilling/internal/webhook/repository"
	"go.uber.org/zap"
)

const (
	statusPending = "PENDING"
)

var ErrMissingTenantID = errors.New("tenant_id required")

type Service struct {
	repo   repository.Repository
	logger *zap.Logger
}

func NewService(repo repository.Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("webhook.dispatch")}
}

func (s *Service) DispatchForEvent(ctx context.Context, env events.EventEnvelope) error {
	if env.TenantID == "" {
		return ErrMissingTenantID
	}
	if env.Subject == "" {
		s.logger.Debug("skipping webhook dispatch without subject", zap.String("event_id", envelopeID(env)))
		return nil
	}

	payload := env.Payload
	if len(payload) == 0 && env.Event != nil {
		var err error
		payload, err = events.MarshalEvent(env.Event)
		if err != nil {
			return err
		}
	}
	if len(payload) == 0 {
		s.logger.Warn("event payload empty, skipping webhook delivery", zap.String("tenant_id", env.TenantID))
		return nil
	}

	webhooks, err := s.repo.ListWebhooksByTenant(ctx, env.TenantID, env.Subject)
	if err != nil {
		return err
	}
	if len(webhooks) == 0 {
		return nil
	}

	now := time.Now().UTC()
	ts := pgtype.Timestamptz{
		Time:  now,
		Valid: true,
	}
	eventID := envelopeID(env)
	for _, hook := range webhooks {
		attempt := repository.DeliveryAttemptParams{
			WebhookID: hook.ID,
			EventID:   eventID,
			TenantID:  env.TenantID,
			Payload:   json.RawMessage(payload),
			Status:    statusPending,
			AttemptNo: 0,
			NextRunAt: ts,
			CreatedAt: ts,
			UpdatedAt: ts,
		}
		if _, err := s.repo.InsertDeliveryAttempt(ctx, attempt); err != nil {
			s.logger.Error("failed to enqueue webhook attempt", zap.String("webhook_id", hook.ID), zap.String("event_id", env.Event.GetId()), zap.Error(err))
			continue
		}
	}
	return nil
}

func envelopeID(env events.EventEnvelope) string {
	if env.Event != nil && env.Event.GetId() != "" {
		return env.Event.GetId()
	}
	return parseEventID(env.Payload)
}

func parseEventID(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	var meta struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(payload, &meta)
	return meta.ID
}
