package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/selvod/selvod/store"
)

// ScopedAuth manages access based on:
// - adminKey: Full access to all endpoints.
// - playbackKey: Access only to the /stream endpoint for minting playback URLs.
// - Library Key: Access to /stream for videos within the key's designated library.
func ScopedAuth(adminKey, playbackKey string, db store.Store) func(http.Handler) http.Handler {
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
			if strings.HasSuffix(r.URL.Path, "/stream") {
				if secureCompare(providedKey, playbackKey) {
					next.ServeHTTP(w, r)
					return
				}

				// Scoped library playback check:
				// Extract video ID from path: /api/v1/videos/{id}/stream
				pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
				if len(pathParts) >= 4 && pathParts[2] == "videos" && pathParts[4] == "stream" {
					videoID := pathParts[3]
					v, err := db.Get(r.Context(), videoID)
					if err == nil && v != nil {
						keys, err := db.ListLibraryKeys(r.Context(), v.LibraryID)
						if err == nil {
							for _, k := range keys {
								if k.IsActive && secureCompare(providedKey, k.PlaybackSecret) {
									next.ServeHTTP(w, r)
									return
								}
							}
						}
					}
				}
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
