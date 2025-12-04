package domain

import "time"

// BillingRun captures a billing execution for a subscription.
type BillingRun struct {
	ID             int64
	TenantID       int64
	SubscriptionID int64
	PeriodStart    time.Time
	PeriodEnd      time.Time
	Status         int32
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
