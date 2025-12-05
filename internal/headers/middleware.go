package headers

import (
	"net/http"

	"google.golang.org/grpc/metadata"
)

// AllowedCORSHeaders exposes the headers safe to expose via CORS.
func AllowedCORSHeaders() []string {
	return append([]string(nil), AllHTTPHeaders...)
}

// ForwardableHeaders lists headers that can be forwarded to downstream services.
func ForwardableHeaders() []string {
	return append([]string(nil), AllMetadataKeys...)
}

// CopyRelevantHeaders copies known HTTP headers into gRPC metadata.
func CopyRelevantHeaders(src http.Header, dst metadata.MD) {
	if dst == nil {
		return
	}
	for _, header := range AllHTTPHeaders {
		value := src.Get(header)
		if value == "" {
			continue
		}
		dst.Set(mapToMetadata(header), value)
	}
}

func mapToMetadata(header string) string {
	switch header {
	case HeaderAPIKey:
		return MetadataAPIKey
	case HeaderTenantID:
		return MetadataTenantID
	case HeaderIdempotency:
		return MetadataIdempotency
	case HeaderEventID:
		return MetadataEventID
	case HeaderEventType:
		return MetadataEventType
	case HeaderCorrelation:
		return MetadataCorrelation
	case HeaderCausation:
		return MetadataCausation
	default:
		return ""
	}
}
