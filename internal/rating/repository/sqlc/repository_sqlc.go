package sqlc

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/rating/domain"
)

// Repository handles rating persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts rating result.
func (r *Repository) Create(ctx context.Context, rating domain.RatingResult) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO rating_results (id, usage_id, amount_cents, currency, created_at) VALUES ($1,$2,$3,$4,$5)`,
		rating.ID, rating.UsageID, rating.AmountCents, rating.Currency, rating.CreatedAt)
	return err
}

// GetByUsage returns ratings for usage.
func (r *Repository) GetByUsage(ctx context.Context, usageID string) ([]domain.RatingResult, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, usage_id, amount_cents, currency, created_at FROM rating_results WHERE usage_id=$1 ORDER BY created_at DESC`, usageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ratings []domain.RatingResult
	for rows.Next() {
		var rRes domain.RatingResult
		if err := rows.Scan(&rRes.ID, &rRes.UsageID, &rRes.AmountCents, &rRes.Currency, &rRes.CreatedAt); err != nil {
			return nil, err
		}
		ratings = append(ratings, rRes)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ratings, nil
}

var _ domain.Repository = (*Repository)(nil)
