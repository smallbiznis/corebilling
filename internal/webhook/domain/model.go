package domain

import (
	"time"

	webhookv1 "github.com/smallbiznis/go-genproto/smallbiznis/webhook/v1"
)

// WebhookSubscription persists webhook configuration.
type WebhookSubscription struct {
	ID         string
	TenantID   string
	EventTypes []string
	URL        string
	Secret     string
	Status     webhookv1.WebhookStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// WebhookDelivery records delivery attempts.
type WebhookDelivery struct {
	ID             string
	SubscriptionID string
	EventID        string
	Status         webhookv1.DeliveryStatus
	Attempt        int32
	HTTPStatus     int32
	ErrorMessage   string
	SentAt         *time.Time
}
