package middleware

import (
	"net/http"
	"strings"
)

// WebSocketCheckOrigin returns a CheckOrigin function aligned with HTTP CORS allow_origins.
// - Empty allowOrigins: permissive (same as historical behavior; set explicit origins in prod).
// - Contains "*": permissive.
// - Otherwise: requests without Origin are allowed (native clients); browser requests must match an entry (case-insensitive).
func WebSocketCheckOrigin(allowOrigins []string) func(*http.Request) bool {
	return func(r *http.Request) bool {
		if len(allowOrigins) == 0 {
			return true
		}
		for _, o := range allowOrigins {
			if strings.TrimSpace(o) == "*" {
				return true
			}
		}
		origin := strings.TrimSpace(r.Header.Get("Origin"))
		if origin == "" {
			return true
		}
		for _, o := range allowOrigins {
			if strings.EqualFold(origin, strings.TrimSpace(o)) {
				return true
			}
		}
		return false
	}
}
