package domain

import "time"

// Subscription represents a tenant subscription to a pricing plan.
type Subscription struct {
	ID        string
	TenantID  string
	PlanID    string
	StartsAt  time.Time
	EndsAt    *time.Time
	Status    string
	CreatedAt time.Time
}
