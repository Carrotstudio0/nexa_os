package dns

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
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
	}

	// Ensure Default Shortcuts are always present
	now := time.Now().String()
	localIP := utils.GetLocalIP()
	defaults := map[string]*nexa.DNSRecord{
		"test.nexa": {Name: "test.nexa", IP: localIP, Port: 1413, Service: "web", CreatedAt: now},
		"share.n":   {Name: "share.n", IP: localIP, Port: 8081, Service: "storage", CreatedAt: now},
		"admin.n":   {Name: "admin.n", IP: localIP, Port: 8080, Service: "admin", CreatedAt: now},
		"dash.n":    {Name: "dash.n", IP: localIP, Port: 7000, Service: "dashboard", CreatedAt: now},
		"chat.n":    {Name: "chat.n", IP: localIP, Port: 8082, Service: "chat", CreatedAt: now},
	}

	changed := false
	for name, rec := range defaults {
		if _, exists := r.Records[name]; !exists {
			r.Records[name] = rec
			changed = true
		}
	}

	if changed || len(r.Records) == 0 {
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
		utils.LogWarning("DNS", fmt.Sprintf("TLS certificates not found: %v. Using plain TCP.", err))
		// Fall back to plain TCP if TLS fails
		ln, err := net.Listen("tcp", "0.0.0.0:"+config.DNSPort)
		if err != nil {
			utils.LogFatal("DNS", fmt.Sprintf("Failed to listen on port %s: %v", config.DNSPort, err))
			return
		}
		defer ln.Close()
		handleDNSListener(ln)
		return
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	ln, err := tls.Listen("tcp", "0.0.0.0:"+config.DNSPort, tlsConfig)
	if err != nil {
		utils.LogFatal("DNS", fmt.Sprintf("Failed to listen on port %s (TLS): %v", config.DNSPort, err))
		return
	}
	defer ln.Close()
	handleDNSListener(ln)
}

// handleDNSListener accepts and processes DNS connections
func handleDNSListener(ln net.Listener) {
	utils.LogInfo("DNS", fmt.Sprintf("Listening Port:    %s (TCP/TLS)", config.DNSPort))
	utils.SaveEndpoint("dns", fmt.Sprintf("tcp://%s:%s", utils.GetLocalIP(), config.DNSPort))

	// Start DNS Stack
	go startStandardUDPDNS()
	go startZerosmDNS()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleDNS(conn)
	}
}

func startZerosmDNS() {
	// mDNS uses Port 5353 UDP multicast 224.0.0.251
	addr, err := net.ResolveUDPAddr("udp4", "224.0.0.251:5353")
	if err != nil {
		return
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		utils.LogWarning("DNS-PRO", "mDNS Port 5353 busy. Zero-config might be limited.")
		return
	}
	defer conn.Close()

	utils.LogSuccess("DNS-PRO", "Zero-Config mDNS Active (nexa.local)")

	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		// respond to anything asking for nexa or .local
		if n > 12 {
			resp := buildDNSResponse(buffer[:n], conn.LocalAddr().(*net.UDPAddr).IP.String())
			if resp != nil {
				conn.WriteToUDP(resp, remoteAddr)
			}
		}
	}
}

func startStandardUDPDNS() {
	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:53")
	if err != nil {
		utils.LogWarning("DNS-PRO", fmt.Sprintf("Standard DNS (Port 53) failed: %v", err))
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		utils.LogWarning("DNS-PRO", "Port 53 is busy. Standard DNS will not work automatically.")
		return
	}
	defer conn.Close()

	utils.LogSuccess("DNS-PRO", "NEXA Smart DNS Active on Port 53")

	buffer := make([]byte, 512)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		// PROCESS DNS QUERY
		go func(data []byte, addr *net.UDPAddr, interfaceAddr string) {
			response := handleSmartDNSQuery(data, interfaceAddr)
			if response != nil {
				conn.WriteToUDP(response, addr)
			}
		}(buffer[:n], remoteAddr, conn.LocalAddr().(*net.UDPAddr).IP.String())
	}
}

func handleSmartDNSQuery(query []byte, interfaceIP string) []byte {
	if len(query) < 12 {
		return nil
	}

	// Extract domain name from DNS query
	domain := ""
	offset := 12
	for {
		length := int(query[offset])
		if length == 0 {
			break
		}
		if domain != "" {
			domain += "."
		}
		domain += string(query[offset+1 : offset+1+length])
		offset += length + 1
	}

	// IF domain ends with .n or is a nexa system domain
	if strings.HasSuffix(domain, ".n") || strings.HasSuffix(domain, ".nexa") {
		return buildDNSResponse(query, interfaceIP) // Point to local IP
	}

	// ELSE: Forward to Global DNS (8.8.8.8) - Recursive Proxy Mode
	return forwardDNSQuery(query)
}

func forwardDNSQuery(query []byte) []byte {
	dnsServer := "8.8.8.8:53"
	conn, err := net.Dial("udp", dnsServer)
	if err != nil {
		return nil
	}
	defer conn.Close()

	_, err = conn.Write(query)
	if err != nil {
		return nil
	}

	resp := make([]byte, 512)
	n, err := conn.Read(resp)
	if err != nil {
		return nil
	}

	return resp[:n]
}

func buildDNSResponse(query []byte, interfaceIP string) []byte {
	tid := query[0:2]
	flags := []byte{0x81, 0x80} // Standard Answer
	qdcount := query[4:6]

	resp := append(tid, flags...)
	resp = append(resp, qdcount...)                        // Questions
	resp = append(resp, qdcount...)                        // Answers
	resp = append(resp, []byte{0x00, 0x00, 0x00, 0x00}...) // Authority + Additional

	// Extract Question part
	offset := 12
	for {
		length := int(query[offset])
		if length == 0 {
			break
		}
		offset += length + 1
	}
	resp = append(resp, query[12:offset+5]...)

	// Answer Resource Record
	resp = append(resp, []byte{0xc0, 0x0c}...)
	resp = append(resp, []byte{0x00, 0x01, 0x00, 0x01}...)
	resp = append(resp, []byte{0x00, 0x00, 0x00, 0x3c}...)
	resp = append(resp, []byte{0x00, 0x04}...)

	// Return the IP of the interface that received the query
	localIP := net.ParseIP(interfaceIP).To4()
	if localIP == nil {
		localIP = net.ParseIP(utils.GetLocalIP()).To4()
	}
	resp = append(resp, localIP...)
	return resp
}

// Resolve returns the IP and Port for a given name if it exists
func Resolve(name string) (*nexa.DNSRecord, bool) {
	if registry == nil {
		return nil, false
	}
	registry.mu.RLock()
	defer registry.mu.RUnlock()
	rec, exists := registry.Records[name]

	// EXPERT ARCH: Wildcard .n resolution
	if !exists && (strings.HasSuffix(name, ".n") || strings.HasSuffix(name, ".nexa")) {
		return &nexa.DNSRecord{
			Name: name, IP: utils.GetLocalIP(), Port: 8000, Service: "gateway",
		}, true
	}

	return rec, exists
}

// Register adds or updates a record
func Register(name, ip string, port int, service string) error {
	if registry == nil {
		return fmt.Errorf("registry not initialized")
	}
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.Records[name] = &nexa.DNSRecord{
		Name: name, IP: ip, Port: port, Service: service, CreatedAt: time.Now().String(),
	}
	return registry.Save()
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
