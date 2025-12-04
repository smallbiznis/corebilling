package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/usage/domain"
)

// Repository handles usage persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts usage record.
func (r *Repository) Create(ctx context.Context, usage domain.UsageRecord) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO usage_records (id, subscription_id, metric, quantity, recorded_at) VALUES ($1,$2,$3,$4,$5)`,
		usage.ID, usage.SubscriptionID, usage.Metric, usage.Quantity, usage.RecordedAt)
	return err
}

// ListBySubscription lists usage rows.
func (r *Repository) ListBySubscription(ctx context.Context, subscriptionID string) ([]domain.UsageRecord, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, subscription_id, metric, quantity, recorded_at FROM usage_records WHERE subscription_id=$1 ORDER BY recorded_at DESC`, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []domain.UsageRecord
	for rows.Next() {
		var rec domain.UsageRecord
		if err := rows.Scan(&rec.ID, &rec.SubscriptionID, &rec.Metric, &rec.Quantity, &rec.RecordedAt); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

var _ domain.Repository = (*Repository)(nil)
