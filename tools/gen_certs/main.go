package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf("Failed to generate private key: %v\n", err)
		os.Exit(1)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		fmt.Printf("Failed to generate serial number: %v\n", err)
		os.Exit(1)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Nexa Protocol"},
			Country:      []string{"US"},
			Province:     []string{"CA"},
			Locality:     []string{"San Francisco"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("0.0.0.0")},
		DNSNames:              []string{"localhost", "*.nexa"},
	}

	// Self-sign the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		fmt.Printf("Failed to create certificate: %v\n", err)
		os.Exit(1)
	}

	// Ensure certs directory exists
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	certsDir := filepath.Join(execDir, "..", "certs")

	// Also try current directory
	if _, err := os.Stat(certsDir); os.IsNotExist(err) {
		certsDir = "certs"
	}

	if _, err := os.Stat(certsDir); os.IsNotExist(err) {
		os.MkdirAll(certsDir, 0755)
	}

	// Write certificate to file
	certPath := filepath.Join(certsDir, "cert.pem")
	certFile, err := os.Create(certPath)
	if err != nil {
		fmt.Printf("Failed to create cert file: %v\n", err)
		os.Exit(1)
	}
	defer certFile.Close()
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	// Write private key to file
	keyPath := filepath.Join(certsDir, "key.pem")
	keyFile, err := os.Create(keyPath)
	if err != nil {
		fmt.Printf("Failed to create key file: %v\n", err)
		os.Exit(1)
	}
	defer keyFile.Close()
	pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	fmt.Println("✓ TLS Certificates generated successfully!")
	fmt.Printf("  Certificate: %s\n", certPath)
	fmt.Printf("  Private Key: %s\n", keyPath)
	fmt.Println("\nCertificates are valid for 1 year and work with localhost")
}
