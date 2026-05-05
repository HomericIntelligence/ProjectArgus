package server

import (
	"fmt"
	"net/http"
)

// securityHeaders adds baseline HTTP security headers to every response.
// The frame-src directive includes the configured Grafana URL so that
// the /grafana panel matrix iframes are permitted by CSP.
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			fmt.Sprintf(
				"default-src 'self'; script-src 'self' https://unpkg.com; connect-src 'self'; style-src 'self'; img-src 'self' data:; frame-src 'self' %s %s",
				s.cfg.GrafanaURL, s.cfg.LokiURL,
			),
		)
		next.ServeHTTP(w, r)
	})
}
