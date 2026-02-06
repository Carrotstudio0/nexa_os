package gateway

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/services/dns"
	"github.com/MultiX0/nexa/pkg/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	GatewayPort = config.GatewayPort
	AdminTarget = config.AdminTarget
	WebTarget   = config.WebTarget
)

// Global state for messaging
var (
	networkMgr   *network.NetworkManager
	expansionMgr *NetworkExpansionManager
	govManager   *governance.GovernanceManager
)

// GatewayResponse represents the gateway status
type GatewayResponse struct {
	Status       string                 `json:"status"`
	LocalIP      string                 `json:"local_ip"`
	Port         string                 `json:"port"`
	Uptime       time.Duration          `json:"uptime"`
	Services     map[string]interface{} `json:"services"`
	NetworkStats network.NetworkStats   `json:"network_stats"`
}

var startTime time.Time

func init() {
	startTime = time.Now()
}

func Start(nm *network.NetworkManager, gm *governance.GovernanceManager) {
	networkMgr = nm
	govManager = gm
	// Initialize router
	r := chi.NewRouter()

	networkMgr = nm
	expansionMgr = NewNetworkExpansionManager(networkMgr, 9999)
	// Gateway connections tracking
	// Gateway connections tracking
	var connectionCount int64
	var requestCount int64

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt64(&connectionCount, 1)
			atomic.AddInt64(&requestCount, 1)
			defer atomic.AddInt64(&connectionCount, -1)
			next.ServeHTTP(w, r)
		})
	})

	// Metrics reporter
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			reqs := atomic.SwapInt64(&requestCount, 0)
			conns := atomic.LoadInt64(&connectionCount)

			if networkMgr != nil {
				networkMgr.UpdateDeviceMetrics("svc-gateway", network.DeviceMetrics{
					RequestsPerSec: float64(reqs),
					LastActivity:   time.Now().Unix(),
					Custom: map[string]interface{}{
						"active_connections": conns,
					},
				})
				networkMgr.UpdateServiceMetrics("gateway", map[string]interface{}{
					"active_connections": conns,
					"requests_per_sec":   reqs,
				})
			}
		}
	}()

	// Start network expansion
	if err := expansionMgr.Start(); err != nil {
		utils.LogError("Gateway", "Failed to start network expansion", err)
	}

	// Register this gateway as a node
	gatewayIP := utils.GetLocalIP()
	gateway, err := networkMgr.RegisterDevice(
		"gateway-"+gatewayIP,
		"Gateway-"+gatewayIP,
		utils.GetMACAddress(),
		gatewayIP,
		8000,
		network.RoleGateway,
	)
	if err != nil {
		utils.LogError("Network", "Failed to register gateway", err)
	} else {
		utils.LogSuccess("Network", fmt.Sprintf("Gateway registered: %s", gateway.ID))
		go expansionMgr.BroadcastDiscovery(gateway)
	}

	// Start network monitoring
	networkMgr.StartMonitoring()

	// Global Middleware Stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Simple CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	createProxies := func(target string, name string) *httputil.ReverseProxy {
		u, _ := url.Parse(target)
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			utils.LogError("Gateway", fmt.Sprintf("Proxy error to %s", name), err)
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		}
		return proxy
	}

	adminProxy := createProxies(AdminTarget, "Admin")
	webProxy := createProxies(WebTarget, "Storage")
	chatProxy := createProxies(config.ChatTarget, "Chat")
	dashboardProxy := createProxies(config.DashboardTarget, "Dashboard")

	// Smart Host-based Routing (Support for .n and .nexa domains)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := strings.ToLower(strings.Split(r.Host, ":")[0])

			// Universal Domain Matcher: Support .n, .lvh.me, and Magic IP Patterns
			isMagicDomain := strings.HasSuffix(host, ".n") ||
				strings.HasSuffix(host, ".nexa") ||
				strings.HasSuffix(host, ".lvh.me") ||
				strings.Contains(host, ".sslip.io") ||
				strings.Contains(host, ".nip.io")

			if isMagicDomain {
				// Extract primary subdomain (e.g., hub from hub.192.168.1.3.ssl
				subdomain := strings.Split(host, ".")[0]
				switch {
				case subdomain == "admin":
					adminProxy.ServeHTTP(w, r)
					return
				case subdomain == "hub" || subdomain == "dashboard" || host == "nexa.local":
					dashboardProxy.ServeHTTP(w, r)
					return
				case subdomain == "vault" || subdomain == "storage":
					webProxy.ServeHTTP(w, r)
					return
				case subdomain == "chat":
					chatProxy.ServeHTTP(w, r)
					return
				default:
					// UNIVERSAL NEXA PROJECT HOSTING
					projectName := subdomain
					projectPath := filepath.Join("sites", projectName)
					if info, err := os.Stat(projectPath); err == nil && info.IsDir() {
						http.StripPrefix("/", http.FileServer(http.Dir(projectPath))).ServeHTTP(w, r)
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	})

	// API Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/status", handleStatus)
		r.Post("/register-site", handleRegisterSite)

		// Network Expansion Routes
		r.Route("/network", func(r chi.Router) {
			r.Get("/topology", handleNetworkTopology)
			r.Get("/stats", handleNetworkStats)
			r.Get("/devices", handleNetworkDevices)
			r.Post("/relay", handleCreateRelay)
			r.Get("/relay", handleGetRelays)
			r.Delete("/relay/{routeId}", handleDeleteRelay)
			r.Post("/connect/{deviceId}", handleConnectDevice)
			r.Delete("/disconnect/{deviceId}", handleDisconnectDevice)
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	// Service Proxies
	r.Mount("/admin", http.StripPrefix("/admin", adminProxy))
	r.Mount("/storage", http.StripPrefix("/storage", webProxy))
	r.Mount("/chat", http.StripPrefix("/chat", chatProxy))
	r.Mount("/dashboard", http.StripPrefix("/dashboard", dashboardProxy))

	// Project Preview Routes (Speed Access)
	r.Route("/s", func(r chi.Router) {
		r.Get("/{projectName}*", func(w http.ResponseWriter, r *http.Request) {
			projectName := chi.URLParam(r, "projectName")
			projectPath := filepath.Join("sites", projectName)

			if info, err := os.Stat(projectPath); err == nil && info.IsDir() {
				// Special handling for clean paths and index.html
				fs := http.StripPrefix("/s/"+projectName, http.FileServer(http.Dir(projectPath)))
				fs.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Project Not Found in NEXA Sites", 404)
		})
	})

	// Root Gateway Page
	r.Get("/", handleGatewayHome)

	localIP := utils.GetLocalIP()
	// MATRIX PRO: Adaptive Listener (Try Port 80 first)
	addr := ":" + config.GatewayPort
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		utils.LogWarning("Gateway", "Port 80 busy. Using professional fallback Port 8000.")
		addr = ":" + config.GatewayBackup
		ln, err = net.Listen("tcp", addr)
	}

	if err != nil {
		utils.LogError("Gateway", "Failed to start server", err)
		return
	}

	utils.LogSuccess("Gateway", fmt.Sprintf("Matrix Hub Online at http://%s%s", localIP, addr))
	utils.SaveEndpoint("gateway", fmt.Sprintf("http://%s%s", localIP, addr))
	http.Serve(ln, r)
}

// Cleanup

func handleNetworkTopology(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if expansionMgr == nil {
		http.Error(w, "Network manager not initialized", http.StatusServiceUnavailable)
		return
	}
	topology := expansionMgr.GetNetworkTopology()
	json.NewEncoder(w).Encode(topology)
}

func handleNetworkStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if expansionMgr == nil {
		http.Error(w, "Network manager not initialized", http.StatusServiceUnavailable)
		return
	}
	stats := expansionMgr.GetNetworkStats()
	json.NewEncoder(w).Encode(stats)
}

func handleNetworkDevices(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if expansionMgr == nil {
		http.Error(w, "Network manager not initialized", http.StatusServiceUnavailable)
		return
	}
	topology := expansionMgr.GetNetworkTopology()
	json.NewEncoder(w).Encode(topology.Devices)
}

func handleCreateRelay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if expansionMgr == nil {
		http.Error(w, "Network manager not initialized", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		SourceID       string `json:"source_id"`
		TargetID       string `json:"target_id"`
		IntermediateID string `json:"intermediate_id"`
		Priority       int    `json:"priority"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	route, err := expansionMgr.CreateRelayRoute(req.SourceID, req.TargetID, req.IntermediateID, req.Priority)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(route)
}

func handleGetRelays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if expansionMgr == nil {
		http.Error(w, "Network manager not initialized", http.StatusServiceUnavailable)
		return
	}
	routes := expansionMgr.GetRelayRoutes()
	json.NewEncoder(w).Encode(routes)
}

func handleDeleteRelay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if expansionMgr == nil {
		http.Error(w, "Network manager not initialized", http.StatusServiceUnavailable)
		return
	}

	routeID := chi.URLParam(r, "routeId")
	if err := expansionMgr.RemoveRelayRoute(routeID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func handleConnectDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if expansionMgr == nil {
		http.Error(w, "Network manager not initialized", http.StatusServiceUnavailable)
		return
	}

	deviceID := chi.URLParam(r, "deviceId")
	var req struct {
		ConnectionType string `json:"connection_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	connType := network.ConnectionWiFi
	if req.ConnectionType != "" {
		connType = network.ConnectionType(req.ConnectionType)
	}

	if err := networkMgr.ConnectDevice(deviceID, connType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	device := networkMgr.GetDevice(deviceID)
	json.NewEncoder(w).Encode(device)
}

func handleDisconnectDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if expansionMgr == nil {
		http.Error(w, "Network manager not initialized", http.StatusServiceUnavailable)
		return
	}

	deviceID := chi.URLParam(r, "deviceId")
	if err := networkMgr.DisconnectDevice(deviceID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "disconnected"})
}

// Handlers
func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status": "online",
		"ip":     utils.GetLocalIP(),
		"port":   GatewayPort,
		"uptime": time.Since(startTime).Seconds(),
		"services": map[string]string{
			"admin":   AdminTarget,
			"storage": WebTarget,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func handleRegisterSite(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		IP      string `json:"ip"`
		Port    int    `json:"port"`
		Service string `json:"service"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if !strings.HasSuffix(req.Name, ".n") && !strings.HasSuffix(req.Name, ".nexa") {
		req.Name += ".n"
	}
	if err := dns.Register(req.Name, req.IP, req.Port, req.Service); err != nil {
		http.Error(w, "Failed to register site: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered", "domain": req.Name})
}

func handleGatewayHome(w http.ResponseWriter, r *http.Request) {
	// Dynamically discover projects in sites folder
	var projects []string
	entries, _ := os.ReadDir("sites")
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"LocalIP":  utils.GetLocalIP(),
		"Port":     GatewayPort,
		"Uptime":   int(time.Since(startTime).Seconds()),
		"Services": config.Services,
		"Projects": projects,
	}
	tmpl, err := template.New("gateway").Parse(gatewayHTML)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

const gatewayHTML = `
<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NEXA Matrix Gateway</title>
    <script>
        // Force HTTP to avoid ERR_SSL_PROTOCOL_ERROR on mobile devices
        if (window.location.protocol === 'https:') {
            window.location.protocol = 'http:';
        }
    </script>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;800&family=Cairo:wght@400;600;700;900&display=swap" rel="stylesheet">
    <style>
        :root {
            --primary: #6366f1;
            --primary-dark: #4f46e5;
            --secondary: #ec4899;
            --accent: #06b6d4;
            --bg: #0f172a;
            --card-bg: rgba(30, 41, 59, 0.7);
            --glass: rgba(255, 255, 255, 0.05);
            --border: rgba(255, 255, 255, 0.1);
            --text: #f8fafc;
            --text-muted: #94a3b8;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body { 
            font-family: 'Outfit', 'Cairo', sans-serif; 
            background: var(--bg);
            background-image: 
                radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.15) 0px, transparent 50%),
                radial-gradient(circle at 10% 20%, rgba(99, 102, 241, 0.15) 0%, transparent 40%),
                radial-gradient(circle at 90% 80%, rgba(168, 85, 247, 0.15) 0%, transparent 40%);
            color: var(--text);
            font-family: 'Outfit', 'Cairo', sans-serif;
            min-height: 100vh;
            display: flex;
            flex-direction: column;
            overflow-x: hidden;
        }

        .header {
            padding: 40px 20px;
            text-align: center;
            animation: fadeInDown 1s ease;
        }

        .logo-main {
            font-size: 3.5rem;
            font-weight: 800;
            background: var(--primary);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            letter-spacing: -2px;
            margin-bottom: 10px;
        }

        .status-badge {
            background: rgba(34, 211, 238, 0.1);
            color: var(--accent);
            padding: 6px 16px;
            border-radius: 20px;
            font-size: 0.9rem;
            border: 1px solid rgba(34, 211, 238, 0.3);
            display: inline-flex;
            align-items: center;
            gap: 8px;
        }

        .status-dot {
            width: 8px;
            height: 8px;
            background: var(--accent);
            border-radius: 50%;
            animation: pulse 2s infinite;
        }

        .container {
            flex: 1;
            padding: 0 20px 40px;
            max-width: 600px;
            margin: 0 auto;
            width: 100%;
        }

        .grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 15px;
        }

        .card {
            background: var(--card);
            backdrop-filter: blur(12px);
            border: 1px solid rgba(255, 255, 255, 0.1);
            border-radius: 24px;
            padding: 25px 15px;
            text-align: center;
            transition: all 0.3s ease;
            text-decoration: none;
            color: white;
            display: flex;
            flex-direction: column;
            align-items: center;
            gap: 12px;
        }

        .card:active { transform: scale(0.95); background: rgba(99, 102, 241, 0.2); }

        .card i {
            font-size: 2rem;
            background: var(--primary);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        .card h3 { font-size: 1.1rem; font-weight: 600; }
        .card p { font-size: 0.75rem; color: #94a3b8; line-height: 1.4; }

        .full-card { grid-column: span 2; }
        .footer { padding: 30px; text-align: center; font-size: 0.8rem; color: #64748b; }

        @keyframes pulse { 0% { opacity: 0.4; } 50% { opacity: 1; } 100% { opacity: 0.4; } }
        @keyframes fadeInDown { from { opacity: 0; transform: translateY(-20px); } to { opacity: 1; transform: translateY(0); } }
    </style>
</head>
<body>
    <div class="header">
        <h1 class="logo-main">NEXA</h1>
        <div class="status-badge">
            <div class="status-dot"></div>
            Matrix Online
        </div>
    </div>

    <div class="container">
        <div class="grid">
            <a href="/dashboard" class="card full-card">
                <i class="fas fa-microchip"></i>
                <h3>Intelligence Hub</h3>
                <p>التحكم والذكاء المركزي</p>
            </a>
            
            <a href="/admin" class="card">
                <i class="fas fa-user-shield"></i>
                <h3>Admin Center</h3>
                <p>إدارة النظام</p>
            </a>

            <a href="/storage" class="card">
                <i class="fas fa-vault"></i>
                <h3>Digital Vault</h3>
                <p>تخزين الملفات</p>
            </a>

            {{range .Projects}}
            <a href="/s/{{.}}" class="card">
                <i class="fas fa-rocket"></i>
                <h3>{{.}}</h3>
                <p>مشروع مستضاف</p>
            </a>
            {{else}}
            <div class="card full-card" style="opacity: 0.5;">
                <p>لا توجد مشاريع في sites حالياً</p>
            </div>
            {{end}}
        </div>
    </div>

    <div class="footer">
        &copy; 2026 NEXA OS v4.0.0-PRO
    </div>
</body>
</html>
`
