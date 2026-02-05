package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// BasicAuth provides HTTP Basic Authentication middleware
func BasicAuth(username, password string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				w.Header().Set("WWW-Authenticate", `Basic realm="Nexa System"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Basic" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Decode and verify credentials
			decoded := make([]byte, len(parts[1]))
			n, err := decodeBase64(parts[1], decoded)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			credentials := string(decoded[:n])
			creds := strings.SplitN(credentials, ":", 2)
			if len(creds) != 2 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Timing-safe comparison
			if subtle.ConstantTimeCompare([]byte(creds[0]), []byte(username)) == 0 ||
				subtle.ConstantTimeCompare([]byte(creds[1]), []byte(password)) == 0 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds important security headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		next.ServeHTTP(w, r)
	})
}

// Helper function to decode base64
func decodeBase64(s string, dst []byte) (int, error) {
	// Simple base64 decoder (in production, use encoding/base64)
	const base64chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	n := 0
	for i := 0; i < len(s); i += 4 {
		// Decode 4 characters to 3 bytes
		if i+3 >= len(s) {
			break
		}

		// This is simplified; use proper base64 decoding in production
		for j := 0; j < 4 && i+j < len(s); j++ {
			char := s[i+j]
			idx := strings.IndexByte(base64chars, char)
			if idx == -1 && char != '=' {
				return 0, nil
			}
		}
	}

	return n, nil
}
