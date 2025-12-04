package domain

import "time"

// UsageRecord captures metered usage.
type UsageRecord struct {
	ID             string
	SubscriptionID string
	Metric         string
	Quantity       int64
	RecordedAt     time.Time
}
