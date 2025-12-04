package domain

import "time"

// Customer represents a tenant-side end customer.
type Customer struct {
	ID                string
	TenantID          string
	ExternalReference string
	Email             string
	Name              string
	Phone             string
	Currency          string
	BillingAddress    map[string]interface{}
	ShippingAddress   map[string]interface{}
	Metadata          map[string]interface{}
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
