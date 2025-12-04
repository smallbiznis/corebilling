package domain

import "time"

// PricingPlan represents a monetization plan.
type PricingPlan struct {
	ID          string
	Name        string
	Description string
	Currency    string
	AmountCents int64
	CreatedAt   time.Time
}
