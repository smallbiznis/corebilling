package domain_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	webhookmock "github.com/smallbiznis/corebilling/internal/mocks/webhook"
	domain "github.com/smallbiznis/corebilling/internal/webhook/domain"
	webhookv1 "github.com/smallbiznis/go-genproto/smallbiznis/webhook/v1"
	"go.uber.org/zap"
)

func TestService_CreateWebhookSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := webhookmock.NewMockRepository(ctrl)
	var capturedSub domain.WebhookSubscription
	repo.EXPECT().CreateSubscription(gomock.Any(), gomock.AssignableToTypeOf(domain.WebhookSubscription{})).DoAndReturn(func(_ context.Context, sub domain.WebhookSubscription) error {
		capturedSub = sub
		return nil
	}).Times(1)

	service := domain.NewService(repo, zap.NewNop())
	req := &webhookv1.CreateWebhookSubscriptionRequest{
		TenantId:   "tenant",
		EventTypes: []string{"order.created", "order.created", "order.updated"},
		Url:        "https://example.com/webhook",
	}
	_, err := service.CreateWebhookSubscription(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateWebhookSubscription returned %v", err)
	}
	if len(capturedSub.EventTypes) != 2 {
		t.Fatalf("event types = %v, want 2 deduped entries", capturedSub.EventTypes)
	}
}

func TestService_TriggerWebhook(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := webhookmock.NewMockRepository(ctrl)
	repo.EXPECT().ListSubscriptions(gomock.Any(), gomock.Any()).Return([]domain.WebhookSubscription{
		{ID: "sub-1", EventTypes: []string{"order.created"}},
		{ID: "sub-2", EventTypes: []string{"invoice.paid"}},
	}, nil).Times(1)
	var capturedDelivery domain.WebhookDelivery
	repo.EXPECT().CreateDelivery(gomock.Any(), gomock.AssignableToTypeOf(domain.WebhookDelivery{})).DoAndReturn(func(_ context.Context, d domain.WebhookDelivery) error {
		capturedDelivery = d
		return nil
	}).Times(1)

	service := domain.NewService(repo, zap.NewNop())
	req := &webhookv1.TriggerWebhookRequest{
		TenantId: "tenant",
		Event: &webhookv1.WebhookEvent{
			EventType: "order.created",
		},
	}
	resp, err := service.TriggerWebhook(context.Background(), req)
	if err != nil {
		t.Fatalf("TriggerWebhook returned %v", err)
	}
	if !resp.Queued {
		t.Fatalf("expected queued true")
	}
	if capturedDelivery.SubscriptionID != "sub-1" {
		t.Fatalf("delivery subscription = %q, want sub-1", capturedDelivery.SubscriptionID)
	}
}
