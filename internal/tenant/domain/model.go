package domain

import (
	"time"

	tenantv1 "github.com/smallbiznis/go-genproto/smallbiznis/tenant/v1"
)

// Tenant describes an account in the core billing platform.
type Tenant struct {
	ID              int64
	ParentID        int64
	Type            tenantv1.TenantType
	Name            string
	Slug            string
	Status          tenantv1.TenantStatus
	DefaultCurrency string
	CountryCode     string
	Metadata        map[string]interface{}
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// ListFilter controls tenant listing behavior.
type ListFilter struct {
	ParentID string
	Limit    int
	Offset   int
}
