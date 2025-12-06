package middleware

import "net/http"

// Middleware is a functional middleware wrapper.
type Middleware func(next http.Handler) http.Handler

// Chain wraps the handler with the provided middleware stack in order.
func Chain(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}
