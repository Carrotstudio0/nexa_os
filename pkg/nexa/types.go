package nexa

import (
	"fmt"
	"os"
)

// General Protocol Constants
const (
	// Protocol Commands
	CMD_PING    = "PING"
	CMD_FETCH   = "FETCH"
	CMD_PUBLISH = "PUBLISH"
	CMD_LIST    = "LIST"
	CMD_AUTH    = "AUTH"

	// DNS Commands
	DNS_PING     = "PING"
	DNS_RESOLVE  = "RESOLVE"
	DNS_REGISTER = "REGISTER"
	DNS_UPDATE   = "UPDATE"
	DNS_DELETE   = "DELETE"
	DNS_LIST     = "LIST"

	// HTTP Status Codes
	STATUS_OK           = 200
	STATUS_CREATED      = 201
	STATUS_BAD_REQ      = 400
	STATUS_UNAUTHORIZED = 401
	STATUS_NOT_FOUND    = 404
	STATUS_SERVER_ERROR = 500
)

// Standard Ports (initialized at runtime)
var (
	PORT_SERVER = getEnv("PORT_SERVER", "1413")
	PORT_DNS    = getEnv("PORT_DNS", "1112")
	PORT_WEB    = getEnv("PORT_WEB", "3000")
)

// Request defines the structure of a Nexa protocol request
// Includes validation for required fields.
type Request struct {
	Command string `json:"command"`
	Target  string `json:"target,omitempty"`
	Body    string `json:"body,omitempty"`
	Token   string `json:"token,omitempty"` // For auth
}

// Validate checks if the Request has the required fields.
func (r *Request) Validate() error {
	if r.Command == "" {
		return fmt.Errorf("command is required")
	}
	return nil
}

// Response defines the structure of a Nexa protocol response
type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Body    string `json:"body,omitempty"`
}

// DNSRecord defines the structure of a DNS entry
// Includes validation for required fields.
type DNSRecord struct {
	Name      string `json:"name"`
	IP        string `json:"ip"`
	Port      int    `json:"port"`
	Service   string `json:"service"`
	Owner     string `json:"owner,omitempty"` // User who owns this record
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Validate checks if the DNSRecord has valid fields.
func (d *DNSRecord) Validate() error {
	if d.Name == "" {
		return fmt.Errorf("name is required")
	}
	if d.IP == "" {
		return fmt.Errorf("IP is required")
	}
	if d.Port <= 0 {
		return fmt.Errorf("port must be greater than 0")
	}
	return nil
}

// getEnv retrieves the value of the environment variable named by the key.
// Logs a warning if the fallback value is used.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	fmt.Printf("Warning: environment variable %s not set, using fallback value %s\n", key, fallback)
	return fallback
}
