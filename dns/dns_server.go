package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// DNS Commands
const (
	DNS_RESOLVE  = "RESOLVE"  // Get IP:PORT for name
	DNS_REGISTER = "REGISTER" // Regsiter new name
	DNS_UPDATE   = "UPDATE"   // Update existing name
	DNS_DELETE   = "DELETE"   // Delete name
	DNS_LIST     = "LIST"     // List all records
	DNS_PING     = "PING"     // Health Check
)

// Resonse Status Codes
const (
	STATUS_OK        = 200
	STATUS_CREATED   = 201
	STATUS_BAD_REQ   = 400
	STATUS_NOT_FOUND = 404
	STATUS_CONFLICT  = 409
	STATUS_ERROR     = 500
)

// dns record represents a dns entry
type DNSRecord struct {
	Name      string
	IP        string
	Port      int
	Service   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// dns registery store all dns records
type DNSRegistry struct {
	mu      sync.RWMutex
	records map[string]*DNSRecord
}

func NewDNSRegistery() *DNSRegistry {
	registery := &DNSRegistry{
		records: make(map[string]*DNSRecord),
	}

	// add default records
	now := time.Now()
	registery.records["test.nexa"] = &DNSRecord{
		Name:      "test.nexa",
		IP:        "127.0.0.1",
		Port:      1412,
		Service:   "web",
		CreatedAt: now,
		UpdatedAt: now,
	}

	registery.records["storage.nexa"] = &DNSRecord{
		Name:      "storage.nexa",
		IP:        "127.0.0.1",
		Port:      1413,
		Service:   "storage",
		CreatedAt: now,
		UpdatedAt: now,
	}

	return registery

}

func (r *DNSRegistry) Resolve(name string) (*DNSRecord, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	record, exists := r.records[name]
	return record, exists
}

func (r *DNSRegistry) Register(record *DNSRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.records[record.Name]; exists {
		return fmt.Errorf("name already exists")
	}

	record.CreatedAt = time.Now()
	record.UpdatedAt = time.Now()
	r.records[record.Name] = record

	return nil
}

func (r *DNSRegistry) Update(record *DNSRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.records[record.Name]
	if !exists {
		return fmt.Errorf("name not found")
	}

	record.CreatedAt = existing.CreatedAt
	record.UpdatedAt = time.Now()
	r.records[record.Name] = record
	return nil
}

func (r *DNSRegistry) Delete(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.records[name]; !exists {
		return fmt.Errorf("name not found")
	}

	delete(r.records, name)
	return nil
}

func (r *DNSRegistry) List() []*DNSRecord {
	r.mu.Lock()
	defer r.mu.Unlock()

	records := make([]*DNSRecord, 0, len(r.records))
	for _, record := range r.records {
		records = append(records, record)
	}

	return records
}

var registry *DNSRegistry

func main() {
	fmt.Println("DNS Server starting on :1112")

	registry = NewDNSRegistery()

	ln, err := net.Listen("tcp", ":1112")

	if err != nil {
		panic(err)
	}

	defer ln.Close()

	fmt.Println("--- DNS Server ready ---")
	fmt.Println("--- Deafult Records ----")
	for _, rec := range registry.List() {
		fmt.Printf("	%s -> %s:%d (%s)\n", rec.Name, rec.IP, rec.Port, rec.Service)
	}

	fmt.Println()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept error:", err)
			continue
		}

		fmt.Printf("DNS query from %s\n", conn.RemoteAddr())
		go handleDNS(conn)
	}
}

func handleDNS(conn net.Conn) {
	defer conn.Close()
	defer fmt.Println(" --- DNS connection closed --- ")

	reader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fmt.Printf("DNS Query: %s\n", line)

		response := processDNSQuery(line)
		conn.Write([]byte(response + "\n"))
		fmt.Printf("DNS Response sent\n")

	}
}

func processDNSQuery(query string) string {
	parts := strings.Fields(query)
	if len(parts) == 0 {
		return formatError(STATUS_BAD_REQ, "Empty query")
	}

	command := strings.ToUpper(parts[0])

	switch command {

	case DNS_PING:
		return formatSuccess(STATUS_OK, "PONG", fmt.Sprintf("DNS Server running. Records: %d", len(registry.records)))

	case DNS_RESOLVE:
		if len(parts) < 2 {
			return formatError(STATUS_BAD_REQ, "Usage: RESOLVE <name>")
		}

		name := parts[1]
		record, exists := registry.Resolve(name)
		if !exists {
			return formatError(STATUS_NOT_FOUND, fmt.Sprintf("Name '%s' not found", name))
		}

		return formatSuccess(STATUS_OK, "RESOLVED",
			fmt.Sprintf("%s:%d|service=%s|ip=%s", record.IP, record.Port, record.Service, record.IP))

	case DNS_REGISTER:
		// REGISTER name ip port service
		if len(parts) < 5 {
			return formatError(STATUS_BAD_REQ, "Usage: REGISTER <name> <ip> <port> <service>")
		}

		name := parts[1]
		ip := parts[2]
		port := 0
		fmt.Sscanf(parts[3], "%d", &port)
		service := parts[4]

		if port <= 0 || port > 65535 {
			return formatError(STATUS_BAD_REQ, "Invalid port number")
		}

		record := &DNSRecord{
			Name:    name,
			IP:      ip,
			Port:    port,
			Service: service,
		}

		err := registry.Register(record)
		if err != nil {
			return formatError(STATUS_CONFLICT, err.Error())
		}

		return formatSuccess(STATUS_CREATED, "REGISTERED", fmt.Sprintf("%s -> %s:%d", name, ip, port))

	case DNS_UPDATE:
		// UPDATE name ip port service
		if len(parts) < 5 {
			return formatError(STATUS_BAD_REQ, "Usage: UPDATE <name> <ip> <port> <service>")
		}

		name := parts[1]
		ip := parts[2]
		port := 0
		fmt.Sscanf(parts[3], "%d", &port)
		service := parts[4]

		record := &DNSRecord{
			Name:    name,
			IP:      ip,
			Port:    port,
			Service: service,
		}

		err := registry.Update(record)
		if err != nil {
			return formatError(STATUS_NOT_FOUND, err.Error())
		}

		return formatSuccess(STATUS_OK, "UPDATED", fmt.Sprintf("%s -> %s:%d", name, ip, port))

	case DNS_DELETE:
		if len(parts) < 2 {
			return formatError(STATUS_BAD_REQ, "Usage: DELETE <name>")
		}

		name := parts[1]
		err := registry.Delete(name)
		if err != nil {
			return formatError(STATUS_NOT_FOUND, err.Error())
		}

		return formatSuccess(STATUS_OK, "DELETED", fmt.Sprintf("Name '%s' removed", name))

	case DNS_LIST:
		records := registry.List()
		if len(records) == 0 {
			return formatSuccess(STATUS_OK, "EMPTY", "No records")
		}

		var result strings.Builder
		for i, rec := range records {
			if i > 0 {
				result.WriteString("|")
			}
			result.WriteString(fmt.Sprintf("%s=%s:%d(%s)", rec.Name, rec.IP, rec.Port, rec.Service))
		}

		return formatSuccess(STATUS_OK, "LIST", result.String())

	default:
		return formatError(STATUS_BAD_REQ, fmt.Sprintf("Unknown command: %s", command))
	}
}

func formatSuccess(code int, message, body string) string {
	return fmt.Sprintf("%d %s %s", code, message, body)
}

func formatError(code int, message string) string {
	return fmt.Sprintf("%d ERROR %s", code, message)
}
