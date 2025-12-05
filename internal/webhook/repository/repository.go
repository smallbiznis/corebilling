package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/smallbiznis/corebilling/internal/webhook/repository/sqlc"
)

type Webhook struct {
	ID         string
	TenantID   string
	TargetUrl  string
	Secret     string
	EventTypes []string
	Enabled    bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type WebhookDeliveryAttempt struct {
	ID        int64
	WebhookID string
	EventID   string
	TenantID  string
	Payload   []byte
	Status    string
	AttemptNo int32
	NextRunAt time.Time
	LastError string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DeliveryAttemptParams struct {
	WebhookID string
	EventID   string
	TenantID  string
	Payload   []byte
	Status    string
	AttemptNo int32
	NextRunAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateWebhookParams = sqlc.InsertWebhookParams
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
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{queries: sqlc.New(pool)}
}

func (r *repository) CreateWebhook(ctx context.Context, arg CreateWebhookParams) (Webhook, error) {
	res, err := r.queries.InsertWebhook(ctx, arg)
	if err != nil {
		return Webhook{}, err
	}
	return toWebhook(res), nil
}

func (r *repository) GetWebhook(ctx context.Context, id string) (Webhook, error) {
	parsedID, err := parseSnowflake(id)
	if err != nil {
		return Webhook{}, err
	}
	res, err := r.queries.GetWebhookByID(ctx, parsedID)
	if err != nil {
		return Webhook{}, err
	}
	return toWebhook(res), nil
}

func (r *repository) ListWebhooksByTenant(ctx context.Context, tenantID, eventType string) ([]Webhook, error) {
	parsedTenantID, err := parseSnowflake(tenantID)
	if err != nil {
		return nil, err
	}
	records, err := r.queries.ListWebhooksByTenant(ctx, sqlc.ListWebhooksByTenantParams{
		TenantID: parsedTenantID,
		Column2:  eventType,
	})
	if err != nil {
		return nil, err
	}
	webhooks := make([]Webhook, 0, len(records))
	for _, rec := range records {
		webhooks = append(webhooks, toWebhook(rec))
	}
	return webhooks, nil
}

func (r *repository) InsertDeliveryAttempt(ctx context.Context, arg DeliveryAttemptParams) (WebhookDeliveryAttempt, error) {
	webhookID, err := parseSnowflake(arg.WebhookID)
	if err != nil {
		return WebhookDeliveryAttempt{}, err
	}
	eventID, err := parseSnowflake(arg.EventID)
	if err != nil {
		return WebhookDeliveryAttempt{}, err
	}
	tenantID, err := parseSnowflake(arg.TenantID)
	if err != nil {
		return WebhookDeliveryAttempt{}, err
	}
	res, err := r.queries.InsertDeliveryAttempt(ctx, sqlc.InsertDeliveryAttemptParams{
		WebhookID: webhookID,
		EventID:   eventID,
		TenantID:  tenantID,
		Payload:   arg.Payload,
		Status:    arg.Status,
		AttemptNo: arg.AttemptNo,
		NextRunAt: toTimestamptz(arg.NextRunAt),
		CreatedAt: toTimestamptz(arg.CreatedAt),
		UpdatedAt: toTimestamptz(arg.UpdatedAt),
	})
	if err != nil {
		return WebhookDeliveryAttempt{}, err
	}
	return toDeliveryAttempt(res), nil
}

func (r *repository) UpdateDeliveryAttemptStatus(ctx context.Context, arg UpdateDeliveryAttemptStatusParams) error {
	return r.queries.UpdateDeliveryAttemptStatus(ctx, arg)
}

func (r *repository) ListDueDeliveryAttempts(ctx context.Context, limit int32) ([]WebhookDeliveryAttempt, error) {
	records, err := r.queries.ListDueDeliveryAttempts(ctx, sqlc.ListDueDeliveryAttemptsParams{
		NextRunAt: toTimestamptz(time.Now().UTC()),
		Limit:     limit,
	})
	if err != nil {
		return nil, err
	}
	attempts := make([]WebhookDeliveryAttempt, 0, len(records))
	for _, rec := range records {
		attempts = append(attempts, toDeliveryAttempt(rec))
	}
	return attempts, nil
}

func (r *repository) MoveToDLQ(ctx context.Context, arg MoveToDLQParams) error {
	return r.queries.MoveToDLQ(ctx, arg)
}

func toWebhook(rec sqlc.Webhooks) Webhook {
	return Webhook{
		ID:         formatSnowflake(rec.ID),
		TenantID:   formatSnowflake(rec.TenantID),
		TargetUrl:  rec.TargetUrl,
		Secret:     rec.Secret,
		EventTypes: rec.EventTypes,
		Enabled:    rec.Enabled,
		CreatedAt:  toTime(rec.CreatedAt),
		UpdatedAt:  toTime(rec.UpdatedAt),
	}
}

func toDeliveryAttempt(rec sqlc.WebhookDeliveryAttempts) WebhookDeliveryAttempt {
	return WebhookDeliveryAttempt{
		ID:        rec.ID,
		WebhookID: formatSnowflake(rec.WebhookID),
		EventID:   formatSnowflake(rec.EventID),
		TenantID:  formatSnowflake(rec.TenantID),
		Payload:   []byte(rec.Payload),
		Status:    rec.Status,
		AttemptNo: rec.AttemptNo,
		NextRunAt: toTime(rec.NextRunAt),
		LastError: rec.LastError.String,
		CreatedAt: toTime(rec.CreatedAt),
		UpdatedAt: toTime(rec.UpdatedAt),
	}
}

func parseSnowflake(value string) (int64, error) {
	if value == "" {
		return 0, errors.New("snowflake id required")
	}
	return strconv.ParseInt(value, 10, 64)
}

func formatSnowflake(value int64) string {
	return strconv.FormatInt(value, 10)
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func toTime(ts pgtype.Timestamptz) time.Time {
	if ts.Valid {
		return ts.Time
	}
	return time.Time{}
}
