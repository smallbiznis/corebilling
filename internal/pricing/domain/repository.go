package domain

import "context"

// Repository defines persistence for pricing entities.
type Repository interface {
	CreateProduct(ctx context.Context, p Product) error
	GetProduct(ctx context.Context, id int64) (Product, error)
	ListProducts(ctx context.Context, tenantID int64) ([]Product, error)
	CreatePrice(ctx context.Context, p Price) error
	GetPrice(ctx context.Context, tenantID, id int64) (Price, error)
	ListPrices(ctx context.Context, tenantID, productID int64) ([]Price, error)
	CreatePriceTier(ctx context.Context, t PriceTier) error
	ListPriceTiersByPriceIDs(ctx context.Context, priceIDs []int64) ([]PriceTier, error)
}
