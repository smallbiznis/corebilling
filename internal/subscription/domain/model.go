package domain

import "time"

// Subscription represents a tenant subscription to a pricing plan.
type Subscription struct {
	ID                 string
	TenantID           string
	CustomerID         string
	PriceID            string
	Status             int32
	AutoRenew          bool
	StartAt            time.Time
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	TrialStartAt       *time.Time
	TrialEndAt         *time.Time
	CancelAt           *time.Time
	CanceledAt         *time.Time
	Metadata           map[string]interface{}
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
