package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/smallbiznis/corebilling/internal/webhook/repository/sqlc"
)

type Webhook = sqlc.Webhook
type WebhookDeliveryAttempt = sqlc.WebhookDeliveryAttempt
type CreateWebhookParams = sqlc.InsertWebhookParams
type DeliveryAttemptParams = sqlc.InsertDeliveryAttemptParams
type UpdateDeliveryAttemptStatusParams = sqlc.UpdateDeliveryAttemptStatusParams
type MoveToDLQParams = sqlc.MoveToDLQParams

// Repository provides access to webhook persistence.
type Repository interface {
	CreateWebhook(ctx context.Context, arg CreateWebhookParams) (Webhook, error)
	GetWebhook(ctx context.Context, id string) (Webhook, error)
	ListWebhooksByTenant(ctx context.Context, tenantID, eventType string) ([]Webhook, error)

	InsertDeliveryAttempt(ctx context.Context, arg DeliveryAttemptParams) (WebhookDeliveryAttempt, error)
	UpdateDeliveryAttemptStatus(ctx context.Context, arg UpdateDeliveryAttemptStatusParams) error
	ListDueDeliveryAttempts(ctx context.Context, limit int32) ([]WebhookDeliveryAttempt, error)
	MoveToDLQ(ctx context.Context, arg MoveToDLQParams) error
}

type repository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a new webhook repository backed by sqlc.
func NewRepository(db *sql.DB) Repository {
	return &repository{queries: sqlc.New(db)}
}

func (r *repository) CreateWebhook(ctx context.Context, arg CreateWebhookParams) (Webhook, error) {
	return r.queries.InsertWebhook(ctx, arg)
}

func (r *repository) GetWebhook(ctx context.Context, id string) (Webhook, error) {
	return r.queries.GetWebhookByID(ctx, id)
}

func (r *repository) ListWebhooksByTenant(ctx context.Context, tenantID, eventType string) ([]Webhook, error) {
	return r.queries.ListWebhooksByTenant(ctx, sqlc.ListWebhooksByTenantParams{
		TenantID: tenantID,
		Column2:  eventType,
	})
}

func (r *repository) InsertDeliveryAttempt(ctx context.Context, arg DeliveryAttemptParams) (WebhookDeliveryAttempt, error) {
	return r.queries.InsertDeliveryAttempt(ctx, arg)
}

func (r *repository) UpdateDeliveryAttemptStatus(ctx context.Context, arg UpdateDeliveryAttemptStatusParams) error {
	return r.queries.UpdateDeliveryAttemptStatus(ctx, arg)
}

func (r *repository) ListDueDeliveryAttempts(ctx context.Context, limit int32) ([]WebhookDeliveryAttempt, error) {
	return r.queries.ListDueDeliveryAttempts(ctx, sqlc.ListDueDeliveryAttemptsParams{
		NextRunAt: time.Now().UTC(),
		Limit:     limit,
	})
}

func (r *repository) MoveToDLQ(ctx context.Context, arg MoveToDLQParams) error {
	return r.queries.MoveToDLQ(ctx, arg)
}
