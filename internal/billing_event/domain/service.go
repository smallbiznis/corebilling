package domain

import (
	"context"
	"errors"

	"github.com/smallbiznis/corebilling/internal/events/outbox"
	billingeventv1 "github.com/smallbiznis/go-genproto/smallbiznis/billing_event/v1"
	eventv1 "github.com/smallbiznis/go-genproto/smallbiznis/event/v1"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service exposes billing event APIs backed by the outbox store.
type Service struct {
	repo   outbox.OutboxRepository
	logger *zap.Logger
}

// NewService constructs Service.
func NewService(repo outbox.OutboxRepository, logger *zap.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger.Named("billing_event.service"),
	}
}

// EmitEvent records an billing-led event in the outbox.
func (s *Service) EmitEvent(ctx context.Context, evt *eventv1.Event) (*billingeventv1.BillingEvent, error) {
	if evt == nil {
		return nil, errors.New("event payload required")
	}

	if err := s.repo.InsertOutboxEvent(ctx, &outbox.OutboxEvent{
		Subject:    evt.GetSubject(),
		TenantID:   evt.GetTenantId(),
		ResourceID: extractResourceID(evt),
		Event:      evt,
	}); err != nil {
		s.logger.Error("failed to persist event", zap.Error(err))
		return nil, err
	}

	s.logger.Info("event recorded", zap.String("subject", evt.GetSubject()), zap.String("event_id", evt.GetId()))
	return s.buildBillingEvent(evt), nil
}

// DeliverWebhook currently records the intent and reports it as undelivered.
func (s *Service) DeliverWebhook(ctx context.Context, req *billingeventv1.DeliverWebhookRequest) (*billingeventv1.DeliverWebhookResponse, error) {
	if req == nil {
		return nil, errors.New("request required")
	}
	s.logger.Info("webhook delivery not implemented", zap.String("event_id", req.GetEventId()), zap.String("endpoint_id", req.GetEndpointId()))
	return &billingeventv1.DeliverWebhookResponse{Delivered: false}, nil
}

func (s *Service) buildBillingEvent(evt *eventv1.Event) *billingeventv1.BillingEvent {
	created := evt.GetCreatedAt()
	if created == nil {
		created = timestamppb.Now()
	}

	return &billingeventv1.BillingEvent{
		Id:         evt.GetId(),
		Subject:    evt.GetSubject(),
		TenantId:   evt.GetTenantId(),
		ResourceId: extractResourceID(evt),
		Payload:    evt.GetData(),
		CreatedAt:  created,
	}
}

func extractResourceID(evt *eventv1.Event) string {
	if evt == nil {
		return ""
	}
	if evt.Metadata == nil {
		return ""
	}
	if val, ok := evt.Metadata.AsMap()["resource_id"]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
