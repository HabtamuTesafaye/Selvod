package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// ScopedAuth manages access based on two independent keys:
// - adminKey: Full access to all endpoints.
// - playbackKey: Access only to the /stream endpoint for minting playback URLs.
func ScopedAuth(adminKey, playbackKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
				return
			}

			providedKey := parts[1]

			// Admin has full access.
			if secureCompare(providedKey, adminKey) {
				next.ServeHTTP(w, r)
				return
			}

			// Playback key only allowed for the signing endpoint.
			if strings.HasSuffix(r.URL.Path, "/stream") && secureCompare(providedKey, playbackKey) {
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, "Forbidden: Insufficient Scope", http.StatusForbidden)
		})
	}
}

func secureCompare(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
