package idempotency

import (
	"context"
	"errors"
	"strconv"

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
	parsedTenantID, err := parseSnowflake(tenantID)
	if err != nil {
		return nil, err
	}
	dbRecord, err := r.queries.GetRecord(ctx, sqlc.GetRecordParams{TenantID: parsedTenantID, Key: key})
	if err != nil {
		return nil, err
	}
	return &Record{
		TenantID:    formatSnowflake(dbRecord.TenantID),
		Key:         dbRecord.Key,
		RequestHash: dbRecord.RequestHash,
		Response:    dbRecord.Response,
		Status:      Status(dbRecord.Status),
		CreatedAt:   dbRecord.CreatedAt.Time,
		UpdatedAt:   dbRecord.UpdatedAt.Time,
	}, nil
}

func (r *SQLRepository) InsertProcessing(ctx context.Context, tenantID, key, requestHash string) error {
	parsedTenantID, err := parseSnowflake(tenantID)
	if err != nil {
		return err
	}
	return r.queries.InsertProcessing(ctx, sqlc.InsertProcessingParams{
		TenantID:    parsedTenantID,
		Key:         key,
		RequestHash: requestHash,
	})
}

func (r *SQLRepository) MarkCompleted(ctx context.Context, tenantID, key string, response []byte) error {
	parsedTenantID, err := parseSnowflake(tenantID)
	if err != nil {
		return err
	}
	return r.queries.MarkCompleted(ctx, sqlc.MarkCompletedParams{
		TenantID: parsedTenantID,
		Key:      key,
		Response: response,
	})
}

var _ Repository = (*SQLRepository)(nil)

func parseSnowflake(value string) (int64, error) {
	if value == "" {
		return 0, errors.New("snowflake id required")
	}
	return strconv.ParseInt(value, 10, 64)
}

func formatSnowflake(value int64) string {
	return strconv.FormatInt(value, 10)
}
