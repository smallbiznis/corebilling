package headers

// Header constants expose the HTTP names used across the platform.
const (
	HeaderAPIKey      = "X-API-Key"         // Tenant credentials supplied by API clients.
	HeaderTenantID    = "X-Tenant-Id"       // Tenant identifier enforcing multi-tenancy.
	HeaderUserID      = "X-User-Id"         // User identifier enforcing multi-tenancy.
	HeaderIdempotency = "X-Idempotency-Key" // Idempotency guard for write operations.
	HeaderEventID     = "X-Event-Id"        // Event identifier for tracing correlation.
	HeaderEventType   = "X-Event-Type"      // Event subject used for routing.
	HeaderCorrelation = "X-Correlation-Id"  // Correlation ID shared across services.
	HeaderCausation   = "X-Causation-Id"    // Causation ID linking follow-up events.
	HeaderSignature   = "X-Signature"       // HMAC signature header for webhook payloads.
)

// Metadata keys mirror gRPC expectations.
const (
	MetadataAPIKey      = "x-api-key"
	MetadataTenantID    = "x-tenant-id"
	MetadataUserID      = "x-user-id"
	MetadataIdempotency = "x-idempotency-key"
	MetadataEventID     = "x-event-id"
	MetadataEventType   = "x-event-type"
	MetadataCorrelation = "x-correlation-id"
	MetadataCausation   = "x-causation-id"
	MetadataSignature   = "x-signature"
)

// AllHTTPHeaders contains each recognized HTTP header name.
var AllHTTPHeaders = []string{
	HeaderAPIKey,
	HeaderTenantID,
	HeaderUserID,
	HeaderIdempotency,
	HeaderEventID,
	HeaderEventType,
	HeaderCorrelation,
	HeaderCausation,
	HeaderSignature,
}

// AllMetadataKeys lists every supported gRPC metadata key.
var AllMetadataKeys = []string{
	MetadataAPIKey,
	MetadataTenantID,
	MetadataUserID,
	MetadataIdempotency,
	MetadataEventID,
	MetadataEventType,
	MetadataCorrelation,
	MetadataCausation,
	MetadataSignature,
}
