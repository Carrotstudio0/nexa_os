package utils

import (
	"os"
	"path/filepath"
)

// FindFile searches for a file in common locations relative to the executable
func FindFile(filename string) string {
	// Get executable directory
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	// Paths to check
	paths := []string{
		filename,                               // Current dir
		filepath.Join("..", filename),          // Parent dir
		filepath.Join(execDir, filename),       // Exe dir
		filepath.Join(execDir, "..", filename), // Parent of exe dir
		filepath.Join(execDir, "..", "..", filename),
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return filename // Return original if not found
}

// FindCertFiles locates the TLS certificate pair
func FindCertFiles() (string, string) {
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	certPaths := []string{
		"certs",
		"../certs",
		filepath.Join(execDir, "certs"),
		filepath.Join(execDir, "..", "certs"),
	}

	for _, p := range certPaths {
		certFile := filepath.Join(p, "cert.pem")
		keyFile := filepath.Join(p, "key.pem")
		if _, err := os.Stat(certFile); err == nil {
			if _, err := os.Stat(keyFile); err == nil {
				return certFile, keyFile
			}
		}
	}
	return "certs/cert.pem", "certs/key.pem"
}
