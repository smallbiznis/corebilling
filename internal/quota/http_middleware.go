package quota

import (
	"net/http"

	"github.com/smallbiznis/corebilling/internal/headers"
	"github.com/smallbiznis/corebilling/internal/server/http/middleware"
	"go.uber.org/zap"
)

// NewHTTPMiddleware wires quota checks into ingestion endpoints.
func NewHTTPMiddleware(rl RateLimiter, svc *Service, logger *zap.Logger) middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost && (r.URL.Path == "/v1/usage" || r.URL.Path == "/v1/events") {
				tenantID := r.Header.Get(headers.HeaderTenantID)
				if tenantID != "" {
					if !rl.Allow(tenantID) {
						logger.Warn("http rate limit exceeded", zap.String("tenant_id", tenantID))
						http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
						return
					}
					// if err := svc.CheckQuotaForEvent(r.Context(), tenantID); err != nil {
					// 	if errors.Is(err, ErrQuotaExceeded) {
					// 		http.Error(w, "quota exceeded", http.StatusTooManyRequests)
					// 		return
					// 	}
					// 	http.Error(w, "quota check failed", http.StatusInternalServerError)
					// 	return
					// }
					// if err := svc.IncrementUsage(r.Context(), tenantID, 1); err != nil {
					// 	logger.Error("failed to increment quota usage", zap.Error(err), zap.String("tenant_id", tenantID))
					// }
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
