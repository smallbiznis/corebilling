package domain

import "time"

// Invoice represents a bill issued by a tenant.
type Invoice struct {
	ID             string
	TenantID       string
	CustomerID     string
	SubscriptionID string
	Status         int32
	CurrencyCode   string
	TotalCents     int64
	SubtotalCents  int64
	TaxCents       int64
	InvoiceNumber  string
	IssuedAt       *time.Time
	DueAt          *time.Time
	PaidAt         *time.Time
	Metadata       map[string]interface{}
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
