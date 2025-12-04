package domain

import "time"

// Run represents an invoice generation invocation.
type Run struct {
	ID             string
	TenantID       string
	CustomerID     string
	SubscriptionID string
	InvoiceID      string
	PeriodStart    time.Time
	PeriodEnd      time.Time
	CreatedAt      time.Time
}
