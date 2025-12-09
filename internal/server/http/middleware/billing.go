package middleware

import (
	"context"
	"net/http"

	"github.com/smallbiznis/corebilling/internal/headers"
)

func BillingMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ctx = context.WithValue(ctx, headers.MetadataAPIKey, r.Header.Get(headers.HeaderAPIKey))
			ctx = context.WithValue(ctx, headers.MetadataTenantID, r.Header.Get(headers.HeaderTenantID))
			ctx = context.WithValue(ctx, headers.MetadataIdempotency, r.Header.Get(headers.HeaderIdempotency))
			ctx = context.WithValue(ctx, headers.MetadataEventID, r.Header.Get(headers.HeaderEventID))
			ctx = context.WithValue(ctx, headers.MetadataEventType, r.Header.Get(headers.HeaderEventType))
			ctx = context.WithValue(ctx, headers.MetadataCorrelation, r.Header.Get(headers.HeaderCorrelation))
			ctx = context.WithValue(ctx, headers.HeaderCausation, r.Header.Get(headers.HeaderCausation))
			ctx = context.WithValue(ctx, headers.MetadataSignature, r.Header.Get(headers.HeaderSignature))

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
