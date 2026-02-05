package gateway

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/network"
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

func Start() {
	// Initialize router
	r := chi.NewRouter()

	// Initialize Network Manager
	certFile, keyFile := utils.FindCertFiles()
	tlsConfig := &tls.Config{
		ClientAuth: tls.NoClientCert,
	}
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	connConfig := network.ConnectionConfig{
		ConnectionType:    network.ConnectionWiFi,
		Timeout:           10 * time.Second,
		MaxRetries:        3,
		HeartbeatInterval: 30 * time.Second,
		ReconnectWaitTime: 5 * time.Second,
		TLSConfig:         tlsConfig,
	}

	networkMgr = network.NewNetworkManager(connConfig)
	expansionMgr = NewNetworkExpansionManager(networkMgr, 9999)

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

	// API Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/status", handleStatus)

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
	r.Route("/admin", func(r chi.Router) {
		r.Handle("/*", http.StripPrefix("/admin", adminProxy))
	})
	r.Route("/storage", func(r chi.Router) {
		r.Handle("/*", http.StripPrefix("/storage", webProxy))
	})
	r.Route("/chat", func(r chi.Router) {
		r.Handle("/*", http.StripPrefix("/chat", chatProxy))
	})
	r.Route("/dashboard", func(r chi.Router) {
		r.Handle("/*", http.StripPrefix("/dashboard", dashboardProxy))
	})

	// Root Gateway Page
	r.Get("/", handleGatewayHome)

	localIP := utils.GetLocalIP()
	utils.LogInfo("Gateway", fmt.Sprintf("Public Address:    http://%s:%s", localIP, GatewayPort))
	utils.SaveEndpoint("gateway", fmt.Sprintf("http://%s:%s", localIP, GatewayPort))

	if err := http.ListenAndServe(":"+GatewayPort, r); err != nil {
		utils.LogFatal("Gateway", err.Error())
	}
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

func handleGatewayHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"LocalIP":  utils.GetLocalIP(),
		"Port":     GatewayPort,
		"Uptime":   int(time.Since(startTime).Seconds()),
		"Services": config.Services,
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
    <title>Nexa Ultimate | Network Gateway</title>
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
                radial-gradient(at 100% 100%, rgba(236, 72, 153, 0.15) 0px, transparent 50%);
            color: var(--text);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
            overflow-x: hidden;
        }

        .container { 
            width: 100%;
            max-width: 1000px;
            position: relative;
        }

        .glass-panel {
            background: var(--card-bg);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border: 1px solid var(--border);
            border-radius: 32px;
            padding: 40px;
            box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
            animation: fadeIn 0.8s ease-out;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }

        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 40px;
            padding-bottom: 20px;
            border-bottom: 1px solid var(--border);
        }

        .brand h1 {
            font-size: 2.5rem;
            font-weight: 800;
            background: linear-gradient(to right, #6366f1, #ec4899);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            letter-spacing: -1px;
        }

        .brand p {
            color: var(--text-muted);
            font-weight: 400;
            margin-top: 4px;
        }

        .status-badge {
            background: rgba(34, 197, 94, 0.2);
            color: #4ade80;
            padding: 8px 16px;
            border-radius: 100px;
            font-size: 0.9rem;
            font-weight: 600;
            display: flex;
            align-items: center;
            gap: 8px;
            border: 1px solid rgba(34, 197, 94, 0.3);
        }

        .status-dot {
            width: 8px;
            height: 8px;
            background: #4ade80;
            border-radius: 50%;
            box-shadow: 0 0 10px #4ade80;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0% { transform: scale(1); opacity: 1; }
            50% { transform: scale(1.5); opacity: 0.5; }
            100% { transform: scale(1); opacity: 1; }
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 40px;
        }

        .metric-card {
            background: var(--glass);
            border: 1px solid var(--border);
            padding: 24px;
            border-radius: 24px;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
        }

        .metric-card:hover {
            background: rgba(255, 255, 255, 0.08);
            border-color: var(--primary);
            transform: translateY(-4px);
        }

        .metric-label {
            color: var(--text-muted);
            font-size: 0.85rem;
            text-transform: uppercase;
            letter-spacing: 1px;
            margin-bottom: 8px;
        }

        .metric-value {
            font-size: 1.5rem;
            font-weight: 700;
            color: var(--text);
        }

        .services-section h2 {
            font-size: 1.5rem;
            margin-bottom: 24px;
            display: flex;
            align-items: center;
            gap: 12px;
        }

        .services-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
            gap: 24px;
        }

        .service-link {
            text-decoration: none;
            color: inherit;
        }

        .service-card {
            background: linear-gradient(135deg, rgba(99, 102, 241, 0.1), rgba(236, 72, 153, 0.1));
            border: 1px solid var(--border);
            border-radius: 24px;
            padding: 32px;
            display: flex;
            flex-direction: column;
            gap: 16px;
            height: 100%;
            position: relative;
            overflow: hidden;
            transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
        }

        .service-card::before {
            content: '';
            position: absolute;
            top: 0; right: 0;
            width: 100px; height: 100px;
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            filter: blur(60px);
            opacity: 0;
            transition: opacity 0.4s;
        }

        .service-card:hover {
            border-color: var(--secondary);
            transform: scale(1.02);
            box-shadow: 0 20px 40px -15px rgba(0, 0, 0, 0.4);
        }

        .service-card:hover::before { opacity: 0.3; }

        .service-card h3 {
            font-size: 1.25rem;
            font-weight: 700;
            color: #fff;
        }

        .service-card p {
            color: var(--text-muted);
            font-size: 0.95rem;
            line-height: 1.6;
        }

        .btn-access {
            margin-top: auto;
            background: rgba(255, 255, 255, 0.1);
            color: white;
            padding: 12px 24px;
            border-radius: 12px;
            text-align: center;
            font-weight: 600;
            transition: all 0.3s;
            border: 1px solid var(--border);
        }

        .service-card:hover .btn-access {
            background: linear-gradient(to right, var(--primary), var(--secondary));
            border-color: transparent;
            box-shadow: 0 4px 15px rgba(99, 102, 241, 0.4);
        }

        .footer {
            margin-top: 40px;
            text-align: center;
            color: var(--text-muted);
            font-size: 0.9rem;
        }

        [dir="rtl"] .brand h1 { letter-spacing: 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="glass-panel">
            <div class="header">
                <div class="brand">
                    <h1>NEXA ULTIMATE</h1>
                    <p>المصفوفة المركزية والتحكم في الشبكة</p>
                </div>
                <div class="status-badge">
                    <div class="status-dot"></div>
                    نشط الآن
                </div>
            </div>

            <div class="metrics-grid">
                <div class="metric-card">
                    <div class="metric-label">عنوان الـ IP المحلي</div>
                    <div class="metric-value">{{.LocalIP}}</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">منفذ البوابة</div>
                    <div class="metric-value">{{.Port}}</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">وقت التشغيل</div>
                    <div class="metric-value">{{.Uptime}}s</div>
                </div>
                <div class="metric-card">
                    <div class="metric-label">الأمان</div>
                    <div class="metric-value">TLS 1.3</div>
                </div>
            </div>

            <div class="services-section">
                <h2>⚡ الخدمات المتصلة</h2>
                <div class="services-grid">
                    {{range .Services}}
                    <a href="{{.url}}" class="service-link">
                        <div class="service-card">
                            <h3>{{.name}}</h3>
                            <p>{{.desc}}. جميع الاتصالات مشفرة وآمنة تماماً.</p>
                            <div class="btn-access">دخول النظام ←</div>
                        </div>
                    </a>
                    {{end}}
                </div>
            </div>

            <div class="footer">
                &copy; 2026 Nexa Ultimate System. جميع الحقوق محفوظة.
            </div>
        </div>
    </div>
</body>
</html>
`
