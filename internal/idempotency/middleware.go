package idempotency

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/smallbiznis/corebilling/internal/headers"
)

// Middleware provides grpc-gateway compatible idempotency handling.
func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			key := r.Header.Get(headers.HeaderIdempotency)
			if key == "" {
				key = extractKeyFromBody(body)
			}

			// If no idempotency key is provided, bypass the subsystem.
			if key == "" {
				r.Body = io.NopCloser(bytes.NewReader(body))
				next.ServeHTTP(w, r)
				return
			}

			tenantID := r.Header.Get(headers.HeaderTenantID)
			ctx := WithIdempotencyKey(r.Context(), key)

			record, existing, err := svc.Begin(ctx, tenantID, key, body)
			if err != nil {
				if IsAlreadyCompleted(err) {
					http.Error(w, ErrAlreadyCompleted.Error(), http.StatusConflict)
					return
				}
				http.Error(w, "idempotency initialization failed", http.StatusInternalServerError)
				return
			}

			if existing {
				if record != nil && record.Status == StatusCompleted && len(record.Response) > 0 {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write(record.Response)
					return
				}
				w.WriteHeader(http.StatusAccepted)
				return
			}

			recorder := &responseRecorder{ResponseWriter: w}
			r = r.WithContext(ctx)
			r.Body = io.NopCloser(bytes.NewReader(body))

			next.ServeHTTP(recorder, r)

			payload := json.RawMessage(recorder.body.Bytes())
			if err := svc.Complete(r.Context(), tenantID, key, payload); err != nil {
				http.Error(w, "failed to store idempotent response", http.StatusInternalServerError)
				return
			}
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	body   bytes.Buffer
	status int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func extractKeyFromBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	if val, ok := payload["idempotency_key"]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
