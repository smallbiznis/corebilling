package domain

import "time"

// Product represents a purchasable item.
type Product struct {
    ID          int64
    TenantID    int64
    Name        string
    Code        string
    Description string
    Active      bool
    Metadata    map[string]interface{}
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Price represents a billing price for a product.
type Price struct {
    ID                   int64
    TenantID             int64
    ProductID            int64
    Code                 string
    LookupKey            string
    PricingModel         int32
    Currency             string
    UnitAmountCents      int64
    BillingInterval      int32
    BillingIntervalCount int32
    Active               bool
    Metadata             map[string]interface{}
    CreatedAt            time.Time
    UpdatedAt            time.Time
}
