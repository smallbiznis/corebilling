package domain

import (
	"time"

	auditv1 "github.com/smallbiznis/go-genproto/smallbiznis/audit/v1"
)

// AuditEvent represents a record stored for accountability dashboards.
type AuditEvent struct {
	ID           string
	TenantID     string
	ActorType    auditv1.ActorType
	ActorID      string
	Action       string
	ActionType   auditv1.ActionType
	ResourceType string
	ResourceID   string
	OldValues    map[string]interface{}
	NewValues    map[string]interface{}
	IpAddress    string
	UserAgent    string
	Metadata     map[string]interface{}
	CreatedAt    time.Time
}

// ListFilter describes how audit history should be scoped.
type ListFilter struct {
	TenantID     string
	ResourceType string
	ResourceID   string
	ActorID      string
	ActionType   auditv1.ActionType
	Limit        int
	Offset       int
}
