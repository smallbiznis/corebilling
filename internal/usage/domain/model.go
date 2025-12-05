package domain

import "time"

// UsageRecord captures metered usage.
type UsageRecord struct {
	ID             string
	TenantID       string
	CustomerID     string
	SubscriptionID string
	MeterCode      string
	Value          float64
	RecordedAt     time.Time
	IdempotencyKey string
	Metadata       map[string]interface{}
	State          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
