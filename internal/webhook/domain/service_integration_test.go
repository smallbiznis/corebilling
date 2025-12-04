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

func TestService_TriggerWebhookIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := webhookmock.NewMockRepository(ctrl)
	repo.EXPECT().ListSubscriptions(gomock.Any(), gomock.Any()).Return([]domain.WebhookSubscription{
		{ID: "sub-one", EventTypes: []string{"order.created"}},
		{ID: "sub-two", EventTypes: []string{"order.created", "invoice.paid"}},
	}, nil).Times(1)
	repo.EXPECT().CreateDelivery(gomock.Any(), gomock.AssignableToTypeOf(domain.WebhookDelivery{})).Return(nil).Times(2)

	service := domain.NewService(repo, zap.NewNop())
	req := &webhookv1.TriggerWebhookRequest{
		TenantId: "tenant",
		Event: &webhookv1.WebhookEvent{
			EventType: "order.created",
		},
	}
	resp, err := service.TriggerWebhook(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.Queued {
		t.Fatalf("expected queued true")
	}
	// We cannot inspect internal mock state; ensure call expectations satisfied by Finish() and no error returned.
}
