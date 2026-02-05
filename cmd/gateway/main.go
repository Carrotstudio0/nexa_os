package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/MultiX0/nexa/pkg/network"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	GatewayPort = "8000"
	AdminTarget = "http://127.0.0.1:8080"
	WebTarget   = "http://127.0.0.1:8081"
)

// Global state for messaging
var (
	messages []Message
	msgMu    sync.RWMutex
)

// Message represents a chat message
type Message struct {
	Timestamp string `json:"timestamp"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	Type      string `json:"type"` // "info", "warning", "error"
}

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
	messages = append(messages, Message{
		Timestamp: time.Now().Format("15:04:05"),
		Sender:    "System",
		Content:   "Welcome to Nexa Gateway - Professional Network System 🚀",
		Type:      "info",
	})
}

func main() {
	// Initialize router
	r := chi.NewRouter()

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

	// Proxy setup
	createProxies := func(target string) *httputil.ReverseProxy {
		u, _ := url.Parse(target)
		proxy := httputil.NewSingleHostReverseProxy(u)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy Error to %s: %v", target, err)
			http.Error(w, fmt.Sprintf("Service Unavailable"), http.StatusServiceUnavailable)
		}
		return proxy
	}

	adminProxy := createProxies(AdminTarget)
	webProxy := createProxies(WebTarget)

	// API Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/status", handleStatus)
		r.Post("/chat/send", handleChatSend)
		r.Get("/chat/messages", handleChatMessages)
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

	// Root Gateway Page
	r.Get("/", handleGatewayHome)

	// 404 Handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Route not found",
			"path":  r.URL.Path,
		})
	})

	localIP := getLocalIP()
	fmt.Printf("\n%s\n", "════════════════════════════════════════════")
	fmt.Printf("🌐  Nexa Central Gateway v2.0\n")
	fmt.Printf("%s\n", "════════════════════════════════════════════")
	fmt.Printf("📍 Listen Address: 0.0.0.0:%s\n", GatewayPort)
	fmt.Printf("🔗 Local Network:  http://%s:%s\n", localIP, GatewayPort)
	fmt.Printf("📱 Mobile Access:  http://%s:%s\n", localIP, GatewayPort)
	fmt.Printf("\n  🎯 Routes:\n")
	fmt.Printf("     /           - Gateway Dashboard\n")
	fmt.Printf("     /admin      - Admin Panel\n")
	fmt.Printf("     /storage    - File Manager\n")
	fmt.Printf("     /api/status - System Status\n")
	fmt.Printf("     /api/chat/* - Chat System\n")
	fmt.Printf("%s\n\n", "════════════════════════════════════════════")

	if err := http.ListenAndServe(":"+GatewayPort, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// Helper function to get local IP
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "localhost"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "localhost"
}

// Handlers
func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status": "online",
		"ip":     getLocalIP(),
		"port":   GatewayPort,
		"uptime": time.Since(startTime).Seconds(),
		"services": map[string]string{
			"admin":   AdminTarget,
			"storage": WebTarget,
		},
	}
	json.NewEncoder(w).Encode(response)
}

func handleChatSend(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}

	sender := r.FormValue("sender")
	content := r.FormValue("message")

	if sender == "" || content == "" {
		http.Error(w, "Missing sender or message", http.StatusBadRequest)
		return
	}

	msgMu.Lock()
	messages = append(messages, Message{
		Timestamp: time.Now().Format("15:04:05"),
		Sender:    sender,
		Content:   content,
		Type:      "info",
	})
	msgMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "sent"})
}

func handleChatMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	msgMu.RLock()
	defer msgMu.RUnlock()
	json.NewEncoder(w).Encode(messages)
}

func handleChatClear(w http.ResponseWriter, r *http.Request) {
	msgMu.Lock()
	messages = messages[:0]
	msgMu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}

func handleGatewayHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := map[string]interface{}{
		"LocalIP": getLocalIP(),
		"Port":    GatewayPort,
		"Uptime":  int(time.Since(startTime).Seconds()),
		"Services": []map[string]string{
			{"name": "Admin Panel", "url": "/admin", "port": "8080"},
			{"name": "File Manager", "url": "/storage", "port": "8081"},
			{"name": "Network Stats", "url": "/api/status", "port": GatewayPort},
		},
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
    <title>Nexa Gateway - البوابة الرئيسية</title>
    <link href="https://fonts.googleapis.com/css2?family=Cairo:wght@400;600;700;800&display=swap" rel="stylesheet">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: 'Cairo', sans-serif; 
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container { 
            background: white; 
            border-radius: 20px;
            padding: 40px;
            max-width: 900px;
            width: 100%;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
        }
        .header {
            text-align: center;
            margin-bottom: 40px;
            border-bottom: 3px solid #667eea;
            padding-bottom: 20px;
        }
        .header h1 { color: #2c3e50; font-size: 2.5em; margin-bottom: 10px; }
        .header p { color: #7f8c8d; font-size: 1.1em; }
        .status-box {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 15px;
            margin-bottom: 30px;
            text-align: center;
        }
        .status-box .status-item { display: inline-block; margin: 0 20px; }
        .services {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .service-card {
            background: #f8f9fa;
            border: 2px solid #e0e0e0;
            border-radius: 15px;
            padding: 25px;
            text-align: center;
            transition: all 0.3s;
        }
        .service-card:hover {
            border-color: #667eea;
            transform: translateY(-5px);
            box-shadow: 0 10px 30px rgba(102, 126, 234, 0.2);
        }
        .service-card h3 { color: #2c3e50; margin-bottom: 10px; }
        .service-card p { color: #7f8c8d; margin-bottom: 15px; }
        .service-card a {
            display: inline-block;
            background: #667eea;
            color: white;
            padding: 10px 20px;
            border-radius: 10px;
            text-decoration: none;
        }
        .service-card a:hover { background: #764ba2; }
        .info-section {
            background: #f0f4ff;
            padding: 20px;
            border-radius: 15px;
            border-left: 4px solid #667eea;
            margin-top: 30px;
        }
        .info-section h3 { color: #2c3e50; margin-bottom: 15px; }
        .info-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 15px;
        }
        .info-item {
            background: white;
            padding: 15px;
            border-radius: 10px;
            text-align: center;
        }
        .info-item .label { color: #7f8c8d; font-size: 0.9em; margin-bottom: 5px; }
        .info-item .value { color: #667eea; font-weight: bold; font-size: 1.1em; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🌐 Nexa Gateway</h1>
            <p>البوابة الرئيسية للنظام المتكامل</p>
        </div>

        <div class="status-box">
            <div class="status-item">
                <strong>الحالة:</strong> <span style="color: #4ade80;">● متصل</span>
            </div>
            <div class="status-item">
                <strong>عنوان IP:</strong> {{.LocalIP}}:{{.Port}}
            </div>
            <div class="status-item">
                <strong>النسخة:</strong> 2.0 Pro
            </div>
        </div>

        <div class="services">
            {{range .Services}}
            <div class="service-card">
                <h3>{{.name}}</h3>
                <p>Port: {{.port}}</p>
                <a href="{{.url}}">الدخول →</a>
            </div>
            {{end}}
        </div>

        <div class="info-section">
            <h3>📊 معلومات النظام</h3>
            <div class="info-grid">
                <div class="info-item">
                    <div class="label">المنصة</div>
                    <div class="value">Nexa v2.0</div>
                </div>
                <div class="info-item">
                    <div class="label">الحالة</div>
                    <div class="value">✓ متشغل</div>
                </div>
                <div class="info-item">
                    <div class="label">الأمان</div>
                    <div class="value">TLS 1.3</div>
                </div>
                <div class="info-item">
                    <div class="label">الخدمات</div>
                    <div class="value">3 نشطة</div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>
`
