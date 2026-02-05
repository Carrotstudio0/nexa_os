package dns

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MultiX0/nexa/pkg/audit"
	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/nexa"
	"github.com/MultiX0/nexa/pkg/utils"
)

// Persistent DNS Registry
type DNSRegistry struct {
	mu       sync.RWMutex
	Records  map[string]*nexa.DNSRecord `json:"records"`
	Filename string                     `json:"-"`
}

var (
	registry   *DNSRegistry
	netManager *network.NetworkManager
	govManager *governance.GovernanceManager
	queryCount int64
)

func NewDNSRegistry(filename string) *DNSRegistry {
	r := &DNSRegistry{
		Records:  make(map[string]*nexa.DNSRecord),
		Filename: filename,
	}

	if _, err := os.Stat(filename); err == nil {
		data, err := os.ReadFile(filename)
		if err == nil {
			json.Unmarshal(data, &r.Records)
		}
	} else {
		r.Records["test.nexa"] = &nexa.DNSRecord{
			Name: "test.nexa", IP: "127.0.0.1", Port: 1413, Service: "web", CreatedAt: time.Now().String(),
		}
		r.Save()
	}
	return r
}

func (r *DNSRegistry) Save() error {
	data, err := json.MarshalIndent(r.Records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.Filename, data, 0644)
}

func Start(nm *network.NetworkManager, gm *governance.GovernanceManager) {
	netManager = nm
	govManager = gm
	audit.Init("dns_audit.log")
	registry = NewDNSRegistry("dns_records.json")

	// Metrics reporter
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			if netManager != nil {
				qCount := atomic.SwapInt64(&queryCount, 0)
				netManager.UpdateServiceMetrics("dns", map[string]interface{}{
					"queries_per_sec": float64(qCount) / 2.0,
					"active_records":  len(registry.Records),
					"status":          "Ready",
				})
			}
		}
	}()

	certFile, keyFile := utils.FindCertFiles()
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load TLS certs: %v", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	ln, err := tls.Listen("tcp", "0.0.0.0:"+config.DNSPort, tlsConfig)
	if err != nil {
		panic(err)
	}
	// Removing defer ln.Close() as it should keep running. Actually defer is fine if it wraps the loop.

	utils.LogInfo("DNS", fmt.Sprintf("Listening Port:    %s (TLS)", config.DNSPort))
	utils.SaveEndpoint("dns", fmt.Sprintf("tls://%s:%s", utils.GetLocalIP(), config.DNSPort))

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleDNS(conn)
	}
}

func handleDNS(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	remoteAddr := conn.RemoteAddr().String()

	for {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		response := processDNSQuery(line, remoteAddr)
		atomic.AddInt64(&queryCount, 1)
		conn.Write([]byte(response + "\n"))
	}
}

func processDNSQuery(query, ip string) string {
	parts := strings.Fields(query)
	if len(parts) == 0 {
		return formatError(nexa.STATUS_BAD_REQ, "Empty query")
	}

	command := strings.ToUpper(parts[0])

	switch command {
	case nexa.DNS_PING:
		return formatSuccess(nexa.STATUS_OK, "PONG", fmt.Sprintf("Records: %d", len(registry.Records)))

	case nexa.DNS_RESOLVE:
		if len(parts) < 2 {
			return formatError(nexa.STATUS_BAD_REQ, "Usage: RESOLVE <name>")
		}
		name := parts[1]
		registry.mu.RLock()
		rec, exists := registry.Records[name]
		registry.mu.RUnlock()

		if !exists {
			return formatError(nexa.STATUS_NOT_FOUND, "Name not found")
		}
		return formatSuccess(nexa.STATUS_OK, "RESOLVED", fmt.Sprintf("%s:%d|service=%s", rec.IP, rec.Port, rec.Service))

	case nexa.DNS_REGISTER:
		if len(parts) < 5 {
			return formatError(nexa.STATUS_BAD_REQ, "Usage: REGISTER <name> <ip> <port> <service>")
		}
		port := 0
		fmt.Sscanf(parts[3], "%d", &port)
		name := parts[1]

		registry.mu.Lock()
		registry.Records[name] = &nexa.DNSRecord{
			Name: name, IP: parts[2], Port: port, Service: parts[4], CreatedAt: time.Now().String(),
		}
		err := registry.Save()
		registry.mu.Unlock()

		if err != nil {
			return formatError(nexa.STATUS_SERVER_ERROR, "Failed to save record")
		}

		audit.Log("GUEST", "REGISTER", name, "SUCCESS", ip)
		return formatSuccess(nexa.STATUS_CREATED, "REGISTERED", name)

	case nexa.DNS_LIST:
		registry.mu.RLock()
		defer registry.mu.RUnlock()
		var list []string
		for k := range registry.Records {
			list = append(list, k)
		}
		return formatSuccess(nexa.STATUS_OK, "LIST", strings.Join(list, ","))

	default:
		return formatError(nexa.STATUS_BAD_REQ, "Unknown Command")
	}
}

func formatSuccess(code int, msg, body string) string {
	return fmt.Sprintf("%d %s %s", code, msg, body)
}

func formatError(code int, msg string) string {
	return fmt.Sprintf("%d ERROR %s", code, msg)
}
