package idempotency

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/idempotency/repository/sqlc"
)

// SQLRepository implements Repository using sqlc-generated queries.
type SQLRepository struct {
	queries *sqlc.Queries
}

// NewSQLRepository constructs a new SQL-backed repository.
func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	return &SQLRepository{queries: sqlc.New(pool)}
}

func (r *SQLRepository) Get(ctx context.Context, tenantID, key string) (*Record, error) {
	dbRecord, err := r.queries.GetRecord(ctx, sqlc.GetRecordParams{TenantID: tenantID, Key: key})
	if err != nil {
		return nil, err
	}
	return &Record{
		TenantID:    dbRecord.TenantID,
		Key:         dbRecord.Key,
		RequestHash: dbRecord.RequestHash,
		Response:    dbRecord.Response,
		Status:      Status(dbRecord.Status),
		CreatedAt:   dbRecord.CreatedAt.Time,
		UpdatedAt:   dbRecord.UpdatedAt.Time,
	}, nil
}

func (r *SQLRepository) InsertProcessing(ctx context.Context, tenantID, key, requestHash string) error {
	return r.queries.InsertProcessing(ctx, sqlc.InsertProcessingParams{
		TenantID:    tenantID,
		Key:         key,
		RequestHash: requestHash,
	})
}

func (r *SQLRepository) MarkCompleted(ctx context.Context, tenantID, key string, response []byte) error {
	return r.queries.MarkCompleted(ctx, sqlc.MarkCompletedParams{
		TenantID: tenantID,
		Key:      key,
		Response: response,
	})
}

var _ Repository = (*SQLRepository)(nil)
