package middleware

import (
	"crypto/subtle"
	"encoding/base64"
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
			if subtle.ConstantTimeCompare([]byte(creds[0]), []byte(username)) == 1 &&
				subtle.ConstantTimeCompare([]byte(creds[1]), []byte(password)) == 1 {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
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

// Helper function to decode base64 - uses standard library
func decodeBase64(s string, dst []byte) (int, error) {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return 0, err
	}
	n := copy(dst, decoded)
	return n, nil
}
