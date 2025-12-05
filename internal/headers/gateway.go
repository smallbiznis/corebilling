package headers

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

// MetadataExtractor converts HTTP headers into gRPC metadata for grpc-gateway.
func MetadataExtractor(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.New(nil)
	if req == nil {
		return md
	}
	for _, header := range AllHTTPHeaders {
		if values, ok := req.Header[NormalizeHeader(header)]; ok && len(values) > 0 {
			key := mapToMetadata(header)
			if key == "" {
				continue
			}
			md.Set(key, values[0])
		}
	}
	// Merge idempotency key from query/body fallback.
	headerVal := md.Get(MetadataIdempotency)
	bodyVal := req.URL.Query().Get("idempotency_key")
	merged := MergeIdempotencyKey(first(headerVal), bodyVal)
	if merged != "" {
		md.Set(MetadataIdempotency, merged)
	}
	normalizeMetadata(md)
	return md
}

func normalizeMetadata(md metadata.MD) {
	for key := range md {
		if !contains(AllMetadataKeys, key) && !strings.HasPrefix(key, "x-") {
			delete(md, key)
		}
	}
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
