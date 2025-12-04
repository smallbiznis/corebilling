package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	webhookv1 "github.com/smallbiznis/go-genproto/smallbiznis/webhook/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Service implements webhookv1.WebhookServiceServer.
type Service struct {
	webhookv1.UnimplementedWebhookServiceServer
	repo   Repository
	logger *zap.Logger
}

// NewService constructs a webhook service.
func NewService(repo Repository, logger *zap.Logger) *Service {
	return &Service{repo: repo, logger: logger.Named("webhook.service")}
}

func (s *Service) CreateWebhookSubscription(ctx context.Context, req *webhookv1.CreateWebhookSubscriptionRequest) (*webhookv1.CreateWebhookSubscriptionResponse, error) {
	if req == nil || req.GetTenantId() == "" || req.GetUrl() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and url are required")
	}
	if len(req.GetEventTypes()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "event_types must contain at least one element")
	}
	now := time.Now().UTC()
	subscription := WebhookSubscription{
		ID:         uuid.NewString(),
		TenantID:   req.GetTenantId(),
		EventTypes: dedupeEventTypes(req.GetEventTypes()),
		URL:        req.GetUrl(),
		Secret:     uuid.NewString(),
		Status:     webhookv1.WebhookStatus_ACTIVE,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.CreateSubscription(ctx, subscription); err != nil {
		s.logger.Error("failed to create webhook subscription", zap.Error(err))
		return nil, err
	}
	return &webhookv1.CreateWebhookSubscriptionResponse{Subscription: s.toSubscriptionProto(subscription)}, nil
}

func (s *Service) ListWebhookSubscriptions(ctx context.Context, req *webhookv1.ListWebhookSubscriptionsRequest) (*webhookv1.ListWebhookSubscriptionsResponse, error) {
	if req == nil || req.GetTenantId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}
	subscriptions, err := s.repo.ListSubscriptions(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}
	resp := &webhookv1.ListWebhookSubscriptionsResponse{}
	for _, item := range subscriptions {
		resp.Subscriptions = append(resp.Subscriptions, s.toSubscriptionProto(item))
	}
	return resp, nil
}

func (s *Service) DeleteWebhookSubscription(ctx context.Context, req *webhookv1.DeleteWebhookSubscriptionRequest) (*webhookv1.DeleteWebhookSubscriptionResponse, error) {
	if req == nil || req.GetTenantId() == "" || req.GetSubscriptionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and subscription_id are required")
	}
	deleted, err := s.repo.DeleteSubscription(ctx, req.GetTenantId(), req.GetSubscriptionId())
	if err != nil {
		return nil, err
	}
	return &webhookv1.DeleteWebhookSubscriptionResponse{Success: deleted}, nil
}

func (s *Service) TriggerWebhook(ctx context.Context, req *webhookv1.TriggerWebhookRequest) (*webhookv1.TriggerWebhookResponse, error) {
	if req == nil || req.GetTenantId() == "" || req.GetEvent() == nil || req.GetEvent().GetEventType() == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id and event with type are required")
	}

	subscriptions, err := s.repo.ListSubscriptions(ctx, req.GetTenantId())
	if err != nil {
		return nil, err
	}
	event := req.GetEvent()
	eventID := event.GetId()
	if eventID == "" {
		eventID = uuid.NewString()
	}

	queued := false
	for _, subscription := range subscriptions {
		if !matchesEventType(subscription.EventTypes, event.GetEventType()) {
			continue
		}
		delivery := WebhookDelivery{
			ID:             uuid.NewString(),
			SubscriptionID: subscription.ID,
			EventID:        eventID,
			Status:         webhookv1.DeliveryStatus_DELIVERY_STATUS_UNSPECIFIED,
			Attempt:        0,
			HTTPStatus:     0,
		}
		if err := s.repo.CreateDelivery(ctx, delivery); err != nil {
			s.logger.Error("failed to record delivery", zap.Error(err))
			continue
		}
		queued = true
	}
	return &webhookv1.TriggerWebhookResponse{Queued: queued}, nil
}

func (s *Service) ListWebhookDeliveries(ctx context.Context, req *webhookv1.ListWebhookDeliveriesRequest) (*webhookv1.ListWebhookDeliveriesResponse, error) {
	if req == nil || req.GetSubscriptionId() == "" {
		return nil, status.Error(codes.InvalidArgument, "subscription_id is required")
	}
	deliveries, err := s.repo.ListDeliveries(ctx, req.GetSubscriptionId())
	if err != nil {
		return nil, err
	}
	resp := &webhookv1.ListWebhookDeliveriesResponse{}
	for _, delivery := range deliveries {
		resp.Deliveries = append(resp.Deliveries, s.toDeliveryProto(delivery))
	}
	return resp, nil
}

func (s *Service) toSubscriptionProto(sub WebhookSubscription) *webhookv1.WebhookSubscription {
	return &webhookv1.WebhookSubscription{
		Id:         sub.ID,
		TenantId:   sub.TenantID,
		EventTypes: sub.EventTypes,
		Url:        sub.URL,
		Secret:     sub.Secret,
		Status:     sub.Status,
		CreatedAt:  timestamppb.New(sub.CreatedAt),
		UpdatedAt:  timestamppb.New(sub.UpdatedAt),
	}
}

func (s *Service) toDeliveryProto(del WebhookDelivery) *webhookv1.WebhookDelivery {
	resp := &webhookv1.WebhookDelivery{
		Id:             del.ID,
		SubscriptionId: del.SubscriptionID,
		EventId:        del.EventID,
		Status:         del.Status,
		Attempt:        del.Attempt,
		HttpStatus:     del.HTTPStatus,
		ErrorMessage:   del.ErrorMessage,
	}
	if del.SentAt != nil {
		resp.SentAt = timestamppb.New(*del.SentAt)
	}
	return resp
}

func matchesEventType(eventTypes []string, eventType string) bool {
	for _, candidate := range eventTypes {
		if strings.EqualFold(candidate, eventType) {
			return true
		}
	}
	return false
}

func dedupeEventTypes(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, val := range values {
		norm := strings.TrimSpace(val)
		if norm == "" {
			continue
		}
		if _, ok := seen[norm]; ok {
			continue
		}
		seen[norm] = struct{}{}
		result = append(result, norm)
	}
	return result
}
