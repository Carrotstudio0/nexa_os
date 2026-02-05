package middleware

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger creates a structured request logger
func RequestLogger(logFile string) func(next http.Handler) http.Handler {
	// Ensure log directory exists
	logDir := filepath.Dir(logFile)
	if logDir != "" {
		os.MkdirAll(logDir, 0755)
	}

	// Create log file
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(fmt.Sprintf("failed to create log file: %v", err))
	}

	logger := log.New(io.MultiWriter(os.Stdout, f), "[NEXA] ", log.LstdFlags)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			clientIP := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				clientIP = xff
			}

			logger.Printf(
				"%s %s %s | Status: %d | Duration: %v | Size: %d bytes | IP: %s",
				r.Method,
				r.RequestURI,
				r.Proto,
				wrapped.Status(),
				duration,
				wrapped.BytesWritten(),
				clientIP,
			)
		})
	}
}

// ErrorLogger logs errors with context
func ErrorLogger(message string, err error) {
	log.Printf("[ERROR] %s: %v", message, err)
}
