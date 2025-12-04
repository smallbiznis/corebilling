package sqlc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/invoice/domain"
)

const defaultInvoicePageSize = 50

// Repository handles invoice persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

// Create inserts invoice.
func (r *Repository) Create(ctx context.Context, inv domain.Invoice) error {
	metadata, err := marshalJSON(inv.Metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO invoices (
			id, tenant_id, customer_id, subscription_id, status,
			currency_code, total_cents, subtotal_cents, tax_cents,
			invoice_number, issued_at, due_at, paid_at,
			metadata, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
	`,
		inv.ID,
		inv.TenantID,
		inv.CustomerID,
		inv.SubscriptionID,
		inv.Status,
		inv.CurrencyCode,
		inv.TotalCents,
		inv.SubtotalCents,
		inv.TaxCents,
		inv.InvoiceNumber,
		inv.IssuedAt,
		inv.DueAt,
		inv.PaidAt,
		metadata,
		inv.CreatedAt,
		inv.UpdatedAt,
	)
	return err
}

// GetByID fetches invoice.
func (r *Repository) GetByID(ctx context.Context, id string) (domain.Invoice, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, customer_id, subscription_id, status,
		       currency_code, total_cents, subtotal_cents, tax_cents,
		       invoice_number, issued_at, due_at, paid_at,
		       metadata, created_at, updated_at
		FROM invoices
		WHERE id=$1
	`, id)

	var inv domain.Invoice
	var metadata []byte
	var issuedAt, dueAt, paidAt *time.Time
	if err := row.Scan(
		&inv.ID,
		&inv.TenantID,
		&inv.CustomerID,
		&inv.SubscriptionID,
		&inv.Status,
		&inv.CurrencyCode,
		&inv.TotalCents,
		&inv.SubtotalCents,
		&inv.TaxCents,
		&inv.InvoiceNumber,
		&issuedAt,
		&dueAt,
		&paidAt,
		&metadata,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	); err != nil {
		return domain.Invoice{}, err
	}

	inv.IssuedAt = issuedAt
	inv.DueAt = dueAt
	inv.PaidAt = paidAt
	inv.Metadata = jsonToMap(metadata)
	return inv, nil
}

// List returns invoices matching the filter.
func (r *Repository) List(ctx context.Context, filter domain.ListInvoicesFilter) ([]domain.Invoice, bool, error) {
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
	if filter.SubscriptionID != "" {
		addClause("subscription_id", filter.SubscriptionID)
	}
	if filter.Status > 0 {
		addClause("status", filter.Status)
	}

	query := `
		SELECT id, tenant_id, customer_id, subscription_id, status,
		       currency_code, total_cents, subtotal_cents, tax_cents,
		       invoice_number, issued_at, due_at, paid_at,
		       metadata, created_at, updated_at
		FROM invoices
	`
	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}
	query += " ORDER BY created_at DESC"

	limit := filter.Limit
	if limit <= 0 {
		limit = defaultInvoicePageSize
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

	var invoices []domain.Invoice
	for rows.Next() {
		var inv domain.Invoice
		var metadata []byte
		var issuedAt, dueAt, paidAt *time.Time
		if err := rows.Scan(
			&inv.ID,
			&inv.TenantID,
			&inv.CustomerID,
			&inv.SubscriptionID,
			&inv.Status,
			&inv.CurrencyCode,
			&inv.TotalCents,
			&inv.SubtotalCents,
			&inv.TaxCents,
			&inv.InvoiceNumber,
			&issuedAt,
			&dueAt,
			&paidAt,
			&metadata,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		); err != nil {
			return nil, false, err
		}
		inv.IssuedAt = issuedAt
		inv.DueAt = dueAt
		inv.PaidAt = paidAt
		inv.Metadata = jsonToMap(metadata)
		invoices = append(invoices, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := false
	if len(invoices) > limit {
		hasMore = true
		invoices = invoices[:limit]
	}
	return invoices, hasMore, nil
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
