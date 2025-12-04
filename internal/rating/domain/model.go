package domain

import "time"

// RatingResult represents calculation of usage into charges.
type RatingResult struct {
	ID          string
	UsageID     string
	AmountCents int64
	Currency    string
	CreatedAt   time.Time
}
