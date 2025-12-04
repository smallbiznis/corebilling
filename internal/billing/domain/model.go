package domain

import "time"

// BillingRecord represents a billed charge for a tenant.
type BillingRecord struct {
	ID          string
	TenantID    string
	AmountCents int64
	CreatedAt   time.Time
}
