package config

import (
	"os"
)

type Config struct {
	Gateway struct {
		Port string
		Host string
	}
	Admin struct {
		Port string
	}
	Storage struct {
		Port string
	}
	Chat struct {
		Port string
	}
	Dashboard struct {
		Port string
	}
	DNS struct {
		Port string
	}
}

var GlobalConfig Config

func init() {
	// Initialize with defaults - eventually load from file
	GlobalConfig.Gateway.Port = getEnv("GATEWAY_PORT", "8000")
	GlobalConfig.Admin.Port = getEnv("ADMIN_PORT", "8080")
	GlobalConfig.Storage.Port = getEnv("STORAGE_PORT", "8081")
	GlobalConfig.Chat.Port = getEnv("CHAT_PORT", "8082")
	GlobalConfig.Dashboard.Port = getEnv("DASHBOARD_PORT", "7000")
	GlobalConfig.DNS.Port = getEnv("DNS_PORT", "1112")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Global System Configuration (Legacy support during migration)
const (
	GatewayPort   = "80"   // High Pro: Native port for .n domains
	GatewayBackup = "8000" // Fallback if 80 is busy
	AdminPort     = "8080"
	WebPort       = "8083"
	StoragePort   = "8081"
	ChatPort      = "8082"
	DashboardPort = "7000"
	ServerPort    = "1413"
	DNSPort       = "1112"

	GatewayTarget   = "http://127.0.0.1:8000"
	AdminTarget     = "http://127.0.0.1:8080"
	WebTarget       = "http://127.0.0.1:8083"
	DashboardTarget = "http://127.0.0.1:7000"
	ChatTarget      = "http://127.0.0.1:8082"
)

// Service Metadata
var Services = []map[string]string{
	{"name": "Admin Center", "url": "/admin", "port": AdminPort, "desc": "Advanced System Administration", "icon": "settings"},
	{"name": "Digital Vault", "url": "/storage", "port": StoragePort, "desc": "Secure Decentralized Storage", "icon": "folder"},
	{"name": "Matrix Chat", "url": "/chat", "port": ChatPort, "desc": "Quantum Encrypted Messaging", "icon": "chat"},
	{"name": "Intelligence Hub", "url": "/dashboard", "port": DashboardPort, "desc": "System-wide Matrix Monitoring", "icon": "dashboard"},
}
