package idempotency

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	idempotencysqlc "github.com/smallbiznis/corebilling/internal/idempotency/repository/sqlc"
	"github.com/sqlc-dev/pqtype"
)

// SQLRepository implements Repository using sqlc-generated queries.
type SQLRepository struct {
	queries *idempotencysqlc.Queries
}

// NewSQLRepository constructs a new SQL-backed repository.
func NewSQLRepository(pool *pgxpool.Pool) *SQLRepository {
	db := stdlib.OpenDB(*pool.Config().ConnConfig)
	return &SQLRepository{queries: idempotencysqlc.New(db)}
}

func (r *SQLRepository) Get(ctx context.Context, tenantID, key string) (*Record, error) {
	dbRecord, err := r.queries.GetRecord(ctx, idempotencysqlc.GetRecordParams{TenantID: tenantID, Key: key})
	if err != nil {
		return nil, err
	}
	return &Record{
		TenantID:    dbRecord.TenantID,
		Key:         dbRecord.Key,
		RequestHash: dbRecord.RequestHash,
		Response:    dbRecord.Response.RawMessage,
		Status:      Status(dbRecord.Status),
		CreatedAt:   dbRecord.CreatedAt,
		UpdatedAt:   dbRecord.UpdatedAt,
	}, nil
}

func (r *SQLRepository) InsertProcessing(ctx context.Context, tenantID, key, requestHash string) error {
	return r.queries.InsertProcessing(ctx, idempotencysqlc.InsertProcessingParams{
		TenantID:    tenantID,
		Key:         key,
		RequestHash: requestHash,
	})
}

func (r *SQLRepository) MarkCompleted(ctx context.Context, tenantID, key string, response []byte) error {
	return r.queries.MarkCompleted(ctx, idempotencysqlc.MarkCompletedParams{
		TenantID: tenantID,
		Key:      key,
		Response: pqtype.NullRawMessage{RawMessage: response, Valid: response != nil},
	})
}

var _ Repository = (*SQLRepository)(nil)
