package domain

import "time"

// Invoice represents a bill sent to a tenant.
type Invoice struct {
	ID                 string
	TenantID           string
	BillingPeriodStart time.Time
	BillingPeriodEnd   time.Time
	TotalCents         int64
	Status             string
	CreatedAt          time.Time
}
