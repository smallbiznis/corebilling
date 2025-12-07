package sqlc

import (
	"context"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/smallbiznis/corebilling/internal/pricing/domain"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) domain.Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateProduct(ctx context.Context, p domain.Product) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO products (id, tenant_id, name, code, description, active, metadata, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		p.ID, p.TenantID, p.Name, p.Code, p.Description, p.Active, p.Metadata, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *Repository) GetProduct(ctx context.Context, id int64) (domain.Product, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, name, code, description, active, metadata, created_at, updated_at FROM products WHERE id=$1`, id)
	var p domain.Product
	if err := row.Scan(&p.ID, &p.TenantID, &p.Name, &p.Code, &p.Description, &p.Active, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return domain.Product{}, err
	}
	return p, nil
}

func (r *Repository) ListProducts(ctx context.Context, tenantID int64) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, name, code, description, active, metadata, created_at, updated_at FROM products WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Product
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Code, &p.Description, &p.Active, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) CreatePrice(ctx context.Context, p domain.Price) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO prices (id, tenant_id, product_id, code, lookup_key, pricing_model, currency, unit_amount_cents, billing_interval, billing_interval_count, active, metadata, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, p.TenantID, p.ProductID, p.Code, p.LookupKey, p.PricingModel, p.Currency, p.UnitAmountCents, p.BillingInterval, p.BillingIntervalCount, p.Active, p.Metadata, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *Repository) CreatePriceTier(ctx context.Context, t domain.PriceTier) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO price_tiers (id, price_id, start_quantity, end_quantity, unit_amount_cents, unit, metadata, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		t.ID, t.PriceID, t.StartQuantity, t.EndQuantity, t.UnitAmountCents, t.Unit, t.Metadata, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *Repository) GetPrice(ctx context.Context, id int64) (domain.Price, error) {
	row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, product_id, code, lookup_key, pricing_model, currency, unit_amount_cents, billing_interval, billing_interval_count, active, metadata, created_at, updated_at FROM prices WHERE id=$1`, id)
	var p domain.Price
	if err := row.Scan(&p.ID, &p.TenantID, &p.ProductID, &p.Code, &p.LookupKey, &p.PricingModel, &p.Currency, &p.UnitAmountCents, &p.BillingInterval, &p.BillingIntervalCount, &p.Active, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return domain.Price{}, err
	}
	return p, nil
}

func (r *Repository) ListPrices(ctx context.Context, tenantID, productID int64) ([]domain.Price, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, product_id, code, lookup_key, pricing_model, currency, unit_amount_cents, billing_interval, billing_interval_count, active, metadata, created_at, updated_at FROM prices WHERE tenant_id=$1 AND product_id=$2 ORDER BY created_at DESC`, tenantID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Price
	for rows.Next() {
		var p domain.Price
		if err := rows.Scan(&p.ID, &p.TenantID, &p.ProductID, &p.Code, &p.LookupKey, &p.PricingModel, &p.Currency, &p.UnitAmountCents, &p.BillingInterval, &p.BillingIntervalCount, &p.Active, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) ListPriceTiersByPriceIDs(ctx context.Context, priceIDs []int64) ([]domain.PriceTier, error) {
	if len(priceIDs) == 0 {
		return nil, nil
	}

	rows, err := r.pool.Query(ctx, `SELECT id, price_id, start_quantity, end_quantity, unit_amount_cents, unit, metadata, created_at, updated_at FROM price_tiers WHERE price_id = ANY($1::BIGINT[]) ORDER BY price_id, start_quantity`, buildInt64Array(priceIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tiers []domain.PriceTier
	for rows.Next() {
		var t domain.PriceTier
		if err := rows.Scan(&t.ID, &t.PriceID, &t.StartQuantity, &t.EndQuantity, &t.UnitAmountCents, &t.Unit, &t.Metadata, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tiers = append(tiers, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tiers, nil
}

func buildInt64Array(ids []int64) string {
	var b strings.Builder
	b.WriteByte('{')
	for i, id := range ids {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatInt(id, 10))
	}
	b.WriteByte('}')
	return b.String()
}

var _ domain.Repository = (*Repository)(nil)
