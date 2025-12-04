package sqlc

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/smallbiznis/corebilling/internal/customer/domain"
)

// Repository implements customer persistence.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository constructs a customer repository.
func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, customer domain.Customer) error {
	billing, err := marshalJSON(customer.BillingAddress)
	if err != nil {
		return err
	}
	shipping, err := marshalJSON(customer.ShippingAddress)
	if err != nil {
		return err
	}
	metadata, err := marshalJSON(customer.Metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `INSERT INTO customers (id, tenant_id, external_reference, email, name, phone, currency, billing_address, shipping_address, metadata, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		customer.ID,
		customer.TenantID,
		customer.ExternalReference,
		customer.Email,
		customer.Name,
		customer.Phone,
		customer.Currency,
		billing,
		shipping,
		metadata,
		customer.CreatedAt,
		customer.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Customer, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, external_reference, email, name, phone, currency, billing_address, shipping_address, metadata, created_at, updated_at FROM customers WHERE id=$1`, id)
	var customer domain.Customer
	var billing, shipping, metadata []byte
	if err := row.Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.ExternalReference,
		&customer.Email,
		&customer.Name,
		&customer.Phone,
		&customer.Currency,
		&billing,
		&shipping,
		&metadata,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	); err != nil {
		return domain.Customer{}, err
	}
	customer.BillingAddress = jsonToMap(billing)
	customer.ShippingAddress = jsonToMap(shipping)
	customer.Metadata = jsonToMap(metadata)
	return customer, nil
}

func (r *Repository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]domain.Customer, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, external_reference, email, name, phone, currency, billing_address, shipping_address, metadata, created_at, updated_at FROM customers WHERE tenant_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []domain.Customer
	for rows.Next() {
		var customer domain.Customer
		var billing, shipping, metadata []byte
		if err := rows.Scan(
			&customer.ID,
			&customer.TenantID,
			&customer.ExternalReference,
			&customer.Email,
			&customer.Name,
			&customer.Phone,
			&customer.Currency,
			&billing,
			&shipping,
			&metadata,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		); err != nil {
			return nil, err
		}
		customer.BillingAddress = jsonToMap(billing)
		customer.ShippingAddress = jsonToMap(shipping)
		customer.Metadata = jsonToMap(metadata)
		customers = append(customers, customer)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *Repository) Update(ctx context.Context, customer domain.Customer) error {
	billing, err := marshalJSON(customer.BillingAddress)
	if err != nil {
		return err
	}
	shipping, err := marshalJSON(customer.ShippingAddress)
	if err != nil {
		return err
	}
	metadata, err := marshalJSON(customer.Metadata)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `UPDATE customers SET tenant_id=$2, external_reference=$3, email=$4, name=$5, phone=$6, currency=$7, billing_address=$8, shipping_address=$9, metadata=$10, updated_at=$11 WHERE id=$1`,
		customer.ID,
		customer.TenantID,
		customer.ExternalReference,
		customer.Email,
		customer.Name,
		customer.Phone,
		customer.Currency,
		billing,
		shipping,
		metadata,
		customer.UpdatedAt,
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
