package sqlc

import (
    "context"

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

func (r *Repository) GetPrice(ctx context.Context, id int64) (domain.Price, error) {
    row := r.pool.QueryRow(ctx, `SELECT id, tenant_id, product_id, code, lookup_key, pricing_model, currency, unit_amount_cents, billing_interval, billing_interval_count, active, metadata, created_at, updated_at FROM prices WHERE id=$1`, id)
    var p domain.Price
    if err := row.Scan(&p.ID, &p.TenantID, &p.ProductID, &p.Code, &p.LookupKey, &p.PricingModel, &p.Currency, &p.UnitAmountCents, &p.BillingInterval, &p.BillingIntervalCount, &p.Active, &p.Metadata, &p.CreatedAt, &p.UpdatedAt); err != nil {
        return domain.Price{}, err
    }
    return p, nil
}

func (r *Repository) ListPrices(ctx context.Context, productID int64) ([]domain.Price, error) {
    rows, err := r.pool.Query(ctx, `SELECT id, tenant_id, product_id, code, lookup_key, pricing_model, currency, unit_amount_cents, billing_interval, billing_interval_count, active, metadata, created_at, updated_at FROM prices WHERE product_id=$1 ORDER BY created_at DESC`, productID)
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

var _ domain.Repository = (*Repository)(nil)
