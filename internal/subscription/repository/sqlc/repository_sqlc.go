package sqlc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/subscription/domain"
)

const defaultSubscriptionPageSize = 25

// Repository provides subscription persistence via pgxpool.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts subscription.
func (r *Repository) Create(ctx context.Context, sub domain.Subscription) error {
	metadata, err := marshalJSON(sub.Metadata)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO subscriptions (
			id, tenant_id, customer_id, price_id, status, auto_renew,
			start_at, current_period_start, current_period_end,
			trial_start_at, trial_end_at, cancel_at, canceled_at,
			metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	`,
		sub.ID,
		sub.TenantID,
		sub.CustomerID,
		sub.PriceID,
		sub.Status,
		sub.AutoRenew,
		sub.StartAt,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.TrialStartAt,
		sub.TrialEndAt,
		sub.CancelAt,
		sub.CanceledAt,
		metadata,
		sub.CreatedAt,
		sub.UpdatedAt,
	)
	return err
}

// GetByID fetches subscription by id.
func (r *Repository) GetByID(ctx context.Context, id string) (domain.Subscription, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, customer_id, price_id, status, auto_renew,
		       start_at, current_period_start, current_period_end,
		       trial_start_at, trial_end_at, cancel_at, canceled_at,
		       metadata, created_at, updated_at
		FROM subscriptions
		WHERE id=$1
	`, id)

	var sub domain.Subscription
	var metadata []byte
	var trialStart, trialEnd, cancelAt, canceledAt *time.Time
	if err := row.Scan(
		&sub.ID,
		&sub.TenantID,
		&sub.CustomerID,
		&sub.PriceID,
		&sub.Status,
		&sub.AutoRenew,
		&sub.StartAt,
		&sub.CurrentPeriodStart,
		&sub.CurrentPeriodEnd,
		&trialStart,
		&trialEnd,
		&cancelAt,
		&canceledAt,
		&metadata,
		&sub.CreatedAt,
		&sub.UpdatedAt,
	); err != nil {
		return domain.Subscription{}, err
	}
	sub.TrialStartAt = trialStart
	sub.TrialEndAt = trialEnd
	sub.CancelAt = cancelAt
	sub.CanceledAt = canceledAt
	sub.Metadata = jsonToMap(metadata)
	return sub, nil
}

// List returns subscriptions matching the filter.
func (r *Repository) List(ctx context.Context, filter domain.ListSubscriptionsFilter) ([]domain.Subscription, bool, error) {
	clauses := []string{}
	args := []any{}

	addClause := func(column string, value any) {
		clauses = append(clauses, fmt.Sprintf("%s=$%d", column, len(args)+1))
		args = append(args, value)
	}

	if filter.TenantID != "" {
		addClause("tenant_id", filter.TenantID)
	}
	if filter.CustomerID != "" {
		addClause("customer_id", filter.CustomerID)
	}

	query := `
		SELECT id, tenant_id, customer_id, price_id, status, auto_renew,
		       start_at, current_period_start, current_period_end,
		       trial_start_at, trial_end_at, cancel_at, canceled_at,
		       metadata, created_at, updated_at
		FROM subscriptions
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = defaultSubscriptionPageSize
	}
	fetchLimit := limit + 1

	args = append(args, fetchLimit)
	query += fmt.Sprintf(" LIMIT $%d", len(args))

	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}
	args = append(args, offset)
	query += fmt.Sprintf(" OFFSET $%d", len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		var metadata []byte
		var trialStart, trialEnd, cancelAt, canceledAt *time.Time
		if err := rows.Scan(
			&sub.ID,
			&sub.TenantID,
			&sub.CustomerID,
			&sub.PriceID,
			&sub.Status,
			&sub.AutoRenew,
			&sub.StartAt,
			&sub.CurrentPeriodStart,
			&sub.CurrentPeriodEnd,
			&trialStart,
			&trialEnd,
			&cancelAt,
			&canceledAt,
			&metadata,
			&sub.CreatedAt,
			&sub.UpdatedAt,
		); err != nil {
			return nil, false, err
		}
		sub.TrialStartAt = trialStart
		sub.TrialEndAt = trialEnd
		sub.CancelAt = cancelAt
		sub.CanceledAt = canceledAt
		sub.Metadata = jsonToMap(metadata)
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(subs) > limit {
		hasMore = true
		subs = subs[:limit]
	}
	return subs, hasMore, nil
}

// Update persists subscription changes.
func (r *Repository) Update(ctx context.Context, sub domain.Subscription) error {
	metadata, err := marshalJSON(sub.Metadata)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE subscriptions SET
			customer_id=$2, price_id=$3, status=$4, auto_renew=$5,
			current_period_start=$6, current_period_end=$7,
			trial_start_at=$8, trial_end_at=$9, cancel_at=$10,
			canceled_at=$11, metadata=$12, updated_at=$13
		WHERE id=$1
	`,
		sub.ID,
		sub.CustomerID,
		sub.PriceID,
		sub.Status,
		sub.AutoRenew,
		sub.CurrentPeriodStart,
		sub.CurrentPeriodEnd,
		sub.TrialStartAt,
		sub.TrialEndAt,
		sub.CancelAt,
		sub.CanceledAt,
		metadata,
		sub.UpdatedAt,
	)
	return err
}

func marshalJSON(value map[string]interface{}) ([]byte, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return json.Marshal(value)
}

func jsonToMap(value []byte) map[string]interface{} {
	if len(value) == 0 {
		return nil
	}
	var data map[string]interface{}
	if err := json.Unmarshal(value, &data); err != nil {
		return nil
	}
	return data
}

var _ domain.Repository = (*Repository)(nil)
