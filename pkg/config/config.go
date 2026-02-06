package config

import (
	"fmt"
	"os" // Added missing import
	"sync"

	"gopkg.in/yaml.v3"
)

// Config represents the global system configuration
type Config struct {
	System struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Environment string `yaml:"environment"`
	} `yaml:"system"`

	Network struct {
		UseLocalhostForce bool `yaml:"use_localhost_force"`
		HotspotEnabled    bool `yaml:"hotspot_enabled"`
	} `yaml:"network"`

	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`

	Services struct {
		Gateway   ServiceConfig `yaml:"gateway"`
		DNS       ServiceConfig `yaml:"dns"`
		Dashboard ServiceConfig `yaml:"dashboard"`
		Admin     ServiceConfig `yaml:"admin"`
		Storage   ServiceConfig `yaml:"storage"`
		Chat      ServiceConfig `yaml:"chat"`
		Web       ServiceConfig `yaml:"web"`
	} `yaml:"services"`

	Paths struct {
		DataDir   string `yaml:"data_dir"`
		StaticDir string `yaml:"static_dir"`
	} `yaml:"paths"`
}

type ServiceConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host,omitempty"`
}

var (
	GlobalConfig *Config
	once         sync.Once
)

// Load initializes the configuration
func Load() (*Config, error) {
	var err error
	once.Do(func() {
		GlobalConfig = &Config{}

		// 1. Load from file
		configPath := findConfigFile()
		if configPath != "" {
			file, e := os.ReadFile(configPath)
			if e == nil {
				if e := yaml.Unmarshal(file, GlobalConfig); e != nil {
					err = fmt.Errorf("failed to parse config.yaml: %v", e)
				}
			}
		}

		// 2. Override with Environment Variables (NEXA_ prefix)
		overrideWithEnv()

		// 3. Set Defaults if missing
		setDefaults()
	})
	return GlobalConfig, err
}

// Get returns the loaded config, causing a panic if not loaded (should call Load in main)
func Get() *Config {
	if GlobalConfig == nil {
		// Try lazy load
		_, err := Load()
		if err != nil {
			panic("Configuration not loaded: " + err.Error())
		}
	}
	return GlobalConfig
}

func findConfigFile() string {
	// Look in current dir, then config/, then parent
	candidates := []string{
		"config.yaml",
		"config/config.yaml",
		"../config.yaml",
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}

func overrideWithEnv() {
	// Example: NEXA_SERVICES_GATEWAY_PORT=9000
	// This is a simple manual override for critical fields.
	// In a full implementation, we would use reflection.

	if port := os.Getenv("NEXA_GATEWAY_PORT"); port != "" {
		// parsing int omitted for brevity in this manual mapping
	}
}

func setDefaults() {
	// Fallback defaults if config file was missing or partial
	if GlobalConfig.Services.Gateway.Port == 0 {
		GlobalConfig.Services.Gateway.Port = 8000
	}
	if GlobalConfig.Services.Dashboard.Port == 0 {
		GlobalConfig.Services.Dashboard.Port = 7000
	}
	if GlobalConfig.Services.Admin.Port == 0 {
		GlobalConfig.Services.Admin.Port = 8080
	}
	if GlobalConfig.Services.Storage.Port == 0 {
		GlobalConfig.Services.Storage.Port = 8081
	}
	if GlobalConfig.Services.Chat.Port == 0 {
		GlobalConfig.Services.Chat.Port = 8082
	}
	if GlobalConfig.Services.DNS.Port == 0 {
		GlobalConfig.Services.DNS.Port = 53
	}
	if GlobalConfig.Services.Web.Port == 0 {
		GlobalConfig.Services.Web.Port = 3000
	}
	if GlobalConfig.Server.Port == 0 {
		GlobalConfig.Server.Port = 1413
	}
	if GlobalConfig.System.Version == "" {
		GlobalConfig.System.Version = "v4.0.0-PRO"
	}
	if GlobalConfig.System.Name == "" {
		GlobalConfig.System.Name = "Nexa Ultimate"
	}
	if GlobalConfig.System.Environment == "" {
		GlobalConfig.System.Environment = "production"
	}
	if GlobalConfig.Server.Host == "" {
		GlobalConfig.Server.Host = "0.0.0.0"
	}
	if GlobalConfig.Services.Gateway.Host == "" {
		GlobalConfig.Services.Gateway.Host = "0.0.0.0"
	}
}

// Legacy constants support (Deprecated)
const (
	GatewayPort   = "8000"
	AdminPort     = "8080"
	WebPort       = "3000"
	StoragePort   = "8081"
	ChatPort      = "8082"
	DashboardPort = "7000"
	ServerPort    = "1413"
	DNSPort       = "1112"
)

// Service Metadata (Dynamic based on config would be better, but keeping simple for now)
var Services = []map[string]string{
	{"name": "Admin Center", "url": "/admin", "desc": "Advanced System Administration", "icon": "settings"},
	{"name": "Digital Vault", "url": "/storage", "desc": "Secure Decentralized Storage", "icon": "folder"},
	{"name": "Matrix Chat", "url": "/chat", "desc": "Quantum Encrypted Messaging", "icon": "chat"},
	{"name": "Intelligence Hub", "url": "/dashboard", "desc": "System-wide Matrix Monitoring", "icon": "dashboard"},
}
