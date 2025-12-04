package domain

import "context"

// Repository defines access to webhook subscriptions and deliveries.
type Repository interface {
	CreateSubscription(ctx context.Context, subscription WebhookSubscription) error
	ListSubscriptions(ctx context.Context, tenantID string) ([]WebhookSubscription, error)
	DeleteSubscription(ctx context.Context, tenantID, id string) (bool, error)
	CreateDelivery(ctx context.Context, delivery WebhookDelivery) error
	ListDeliveries(ctx context.Context, subscriptionID string) ([]WebhookDelivery, error)
}
