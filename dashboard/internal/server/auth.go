package server

import (
	"encoding/base64"
	"net/http"
	"strings"
)

// AuthMode represents the authentication scheme for the Atlas dashboard.
type AuthMode string

const (
	AuthNone   AuthMode = "none"
	AuthBasic  AuthMode = "basic"
	AuthBearer AuthMode = "bearer"
)

// Middleware returns a Chi middleware that enforces the configured auth mode.
//
//   - none: no-op passthrough; all requests are allowed.
//   - basic: validates Authorization: Basic <base64(user:pass)>.
//     On failure returns 401 with WWW-Authenticate: Basic realm="Atlas".
//   - bearer: validates Authorization: Bearer <token> OR ?token=<token>.
//     The query-param fallback exists for EventSource / SSE clients that cannot
//     set custom request headers (Accept: text/event-stream).
//     On failure returns 401.
func Middleware(mode AuthMode, user, pass, bearerToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch mode {
			case AuthBasic:
				if !checkBasic(r, user, pass) {
					w.Header().Set("WWW-Authenticate", `Basic realm="Atlas"`)
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			case AuthBearer:
				if !checkBearer(r, bearerToken) {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			default:
				// AuthNone or unknown: allow through.
			}
			next.ServeHTTP(w, r)
		})
	}
}

// checkBasic validates an HTTP Basic auth header against the expected credentials.
func checkBasic(r *http.Request, user, pass string) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Basic ") {
		return false
	}
	encoded := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		return false
	}
	return parts[0] == user && parts[1] == pass
}

// checkBearer validates a Bearer token from either the Authorization header or
// the ?token= query parameter (for SSE / EventSource compatibility).
func checkBearer(r *http.Request, token string) bool {
	// Check Authorization: Bearer <token> header first.
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		provided := strings.TrimPrefix(authHeader, "Bearer ")
		return provided == token
	}
	// Fall back to ?token= query parameter (EventSource compat).
	if qt := r.URL.Query().Get("token"); qt != "" {
		return qt == token
	}
	return false
}
