package sqlc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/usage/domain"
)

const defaultUsagePageSize = 50

// Repository handles usage persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a usage repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts usage record.
func (r *Repository) Create(ctx context.Context, usage domain.UsageRecord) error {
	metadata, err := marshalJSON(usage.Metadata)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO usage_records (
			id, tenant_id, customer_id, subscription_id,
			meter_code, value, recorded_at, idempotency_key,
			metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`,
		usage.ID,
		usage.TenantID,
		usage.CustomerID,
		usage.SubscriptionID,
		usage.MeterCode,
		usage.Value,
		usage.RecordedAt,
		nullIfEmpty(usage.IdempotencyKey),
		metadata,
		usage.CreatedAt,
		usage.UpdatedAt,
	)
	return err
}

// List returns usage records matching the filter.
func (r *Repository) List(ctx context.Context, filter domain.ListUsageFilter) ([]domain.UsageRecord, bool, error) {
	clauses := []string{}
	args := []any{}

	addClause := func(expr string, value any) {
		clauses = append(clauses, fmt.Sprintf("%s $%d", expr, len(args)+1))
		args = append(args, value)
	}

	if filter.TenantID != "" {
		addClause("tenant_id =", filter.TenantID)
	}
	if filter.SubscriptionID != "" {
		addClause("subscription_id =", filter.SubscriptionID)
	}
	if filter.CustomerID != "" {
		addClause("customer_id =", filter.CustomerID)
	}
	if filter.MeterCode != "" {
		addClause("meter_code =", filter.MeterCode)
	}
	if !filter.From.IsZero() {
		addClause("recorded_at >=", filter.From)
	}
	if !filter.To.IsZero() {
		addClause("recorded_at <=", filter.To)
	}

	query := `
		SELECT id, tenant_id, customer_id, subscription_id,
		       meter_code, value, recorded_at, idempotency_key,
		       metadata, created_at, updated_at
		FROM usage_records
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY recorded_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = defaultUsagePageSize
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

	var records []domain.UsageRecord
	for rows.Next() {
		var usage domain.UsageRecord
		var metadata []byte
		if err := rows.Scan(
			&usage.ID,
			&usage.TenantID,
			&usage.CustomerID,
			&usage.SubscriptionID,
			&usage.MeterCode,
			&usage.Value,
			&usage.RecordedAt,
			&usage.IdempotencyKey,
			&metadata,
			&usage.CreatedAt,
			&usage.UpdatedAt,
		); err != nil {
			return nil, false, err
		}
		usage.Metadata = jsonToMap(metadata)
		records = append(records, usage)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(records) > limit {
		hasMore = true
		records = records[:limit]
	}
	return records, hasMore, nil
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

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

var _ domain.Repository = (*Repository)(nil)
