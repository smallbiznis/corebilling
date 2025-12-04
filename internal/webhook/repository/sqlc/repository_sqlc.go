package sqlc

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	webhookv1 "github.com/smallbiznis/go-genproto/smallbiznis/webhook/v1"

	"github.com/smallbiznis/corebilling/internal/webhook/domain"
)

// Repository stores webhook subscriptions and deliveries.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a webhook repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateSubscription(ctx context.Context, subscription domain.WebhookSubscription) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO webhook_subscriptions (
			id, tenant_id, event_types, url, secret, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, subscription.ID, subscription.TenantID, subscription.EventTypes, subscription.URL, subscription.Secret, int16(subscription.Status), subscription.CreatedAt, subscription.UpdatedAt)
	return err
}

func (r *Repository) ListSubscriptions(ctx context.Context, tenantID string) ([]domain.WebhookSubscription, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, event_types, url, secret, status, created_at, updated_at
		FROM webhook_subscriptions WHERE tenant_id=$1 ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.WebhookSubscription
	for rows.Next() {
		var subscription domain.WebhookSubscription
		var eventTypes []string
		var status int16
		if err := rows.Scan(
			&subscription.ID,
			&subscription.TenantID,
			&eventTypes,
			&subscription.URL,
			&subscription.Secret,
			&status,
			&subscription.CreatedAt,
			&subscription.UpdatedAt,
		); err != nil {
			return nil, err
		}
		subscription.EventTypes = eventTypes
		subscription.Status = webhookv1.WebhookStatus(status)
		subs = append(subs, subscription)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}

func (r *Repository) DeleteSubscription(ctx context.Context, tenantID, id string) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM webhook_subscriptions WHERE tenant_id=$1 AND id=$2
	`, tenantID, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *Repository) CreateDelivery(ctx context.Context, delivery domain.WebhookDelivery) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO webhook_deliveries (
			id, subscription_id, event_id, status, response_status, response_body, delivered_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7)
	`, delivery.ID, delivery.SubscriptionID, delivery.EventID, int16(delivery.Status), delivery.HTTPStatus, delivery.ErrorMessage, delivery.SentAt)
	return err
}

func (r *Repository) ListDeliveries(ctx context.Context, subscriptionID string) ([]domain.WebhookDelivery, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, subscription_id, event_id, status, response_status, response_body, delivered_at
		FROM webhook_deliveries WHERE subscription_id=$1 ORDER BY created_at DESC
	`, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []domain.WebhookDelivery
	for rows.Next() {
		var delivery domain.WebhookDelivery
		var status int16
		var deliveredAt *time.Time
		if err := rows.Scan(
			&delivery.ID,
			&delivery.SubscriptionID,
			&delivery.EventID,
			&status,
			&delivery.HTTPStatus,
			&delivery.ErrorMessage,
			&deliveredAt,
		); err != nil {
			return nil, err
		}
		delivery.SentAt = deliveredAt
		delivery.Status = webhookv1.DeliveryStatus(status)
		deliveries = append(deliveries, delivery)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return deliveries, nil
}

var _ domain.Repository = (*Repository)(nil)
