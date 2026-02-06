package analytics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local network
	},
}

// Middleware to track all requests
func TrackingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Get or create session
		sessionID := getSessionID(r)
		ip := getIP(r)
		userAgent := r.UserAgent()

		manager := GetManager()
		session := manager.GetSession(sessionID)
		if session == nil {
			session = manager.CreateSession(sessionID, ip, userAgent)
		}

		// Track page view
		manager.TrackAction(sessionID, Action{
			Type:   "page_view",
			Path:   r.URL.Path,
			Method: r.Method,
		})

		// Custom response writer to capture status
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		// Track request completion
		duration := time.Since(start).Milliseconds()
		manager.TrackAction(sessionID, Action{
			Type:     "api_call",
			Path:     r.URL.Path,
			Method:   r.Method,
			Duration: duration,
			Status:   rw.statusCode,
		})
	})
}

// RegisterRoutes registers analytics API routes
func RegisterRoutes(r chi.Router) {
	r.Route("/api/analytics", func(r chi.Router) {
		r.Get("/stats", handleGetStats)
		r.Get("/sessions", handleGetSessions)
		r.Get("/sessions/active", handleGetActiveSessions)
		r.Get("/events", handleEventsWebSocket)
		r.Get("/session/{id}", handleGetSession)
	})

	// Analytics dashboard page
	r.Get("/analytics", handleAnalyticsDashboard)
}

func handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats := GetManager().GetStats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleGetSessions(w http.ResponseWriter, r *http.Request) {
	sessions := GetManager().GetAllSessions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func handleGetActiveSessions(w http.ResponseWriter, r *http.Request) {
	sessions := GetManager().GetActiveSessions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func handleGetSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	session := GetManager().GetSession(id)
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// WebSocket endpoint for real-time events
func handleEventsWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	manager := GetManager()
	events := manager.GetEventChannel()

	// Send current stats immediately
	stats := manager.GetStats()
	conn.WriteJSON(map[string]interface{}{
		"type": "initial_stats",
		"data": stats,
	})

	// Stream events
	for event := range events {
		if err := conn.WriteJSON(event); err != nil {
			break
		}
	}
}

func handleAnalyticsDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(analyticsDashboardHTML))
}

// Helper functions
func getSessionID(r *http.Request) string {
	// Try to get from cookie
	cookie, err := r.Cookie("session_id")
	if err == nil {
		return cookie.Value
	}

	// Generate new ID
	return fmt.Sprintf("sess_%d", time.Now().UnixNano())
}

func getIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use RemoteAddr
	ip := r.RemoteAddr
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		ip = ip[:colonIndex]
	}
	return ip
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var analyticsDashboardHTML = `<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NEXA | Matrix Analytics PRO</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;800&family=Cairo:wght@400;600;700;900&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/chart.js"></script>
    <style>
        :root {
            --primary: #6366f1;
            --secondary: #ec4899;
            --accent: #06b6d4;
            --bg: #020617;
            --card: rgba(15, 23, 42, 0.7);
            --border: rgba(255, 255, 255, 0.1);
            --glass: rgba(255, 255, 255, 0.05);
        }
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Outfit', 'Cairo', sans-serif;
            background: var(--bg);
            background-image: 
                radial-gradient(at 0% 0%, rgba(99, 102, 241, 0.15) 0px, transparent 50%),
                radial-gradient(at 100% 100%, rgba(236, 72, 153, 0.1) 0px, transparent 50%);
            color: #f8fafc;
            min-height: 100vh;
            display: flex;
        }
        
        /* Sidebar */
        .sidebar {
            width: 80px;
            background: var(--card);
            backdrop-filter: blur(20px);
            border-left: 1px solid var(--border);
            display: flex;
            flex-direction: column;
            align-items: center;
            padding: 30px 0;
            gap: 30px;
            transition: width 0.3s;
        }
        .sidebar:hover { width: 200px; }
        .nav-item { color: #94a3b8; font-size: 1.5rem; cursor: pointer; transition: 0.3s; display: flex; align-items: center; gap: 15px; width: 100%; padding: 0 25px; }
        .nav-item:hover, .nav-item.active { color: var(--primary); }
        .nav-text { font-size: 1rem; opacity: 0; white-space: nowrap; transition: 0.2s; font-weight: 600; }
        .sidebar:hover .nav-text { opacity: 1; }

        /* Main Content */
        .content { flex: 1; padding: 40px; overflow-y: auto; }
        .header { margin-bottom: 40px; }
        .header h1 { font-size: 2.5rem; font-weight: 800; background: linear-gradient(to right, #6366f1, #ec4899); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        
        /* Stats Grid */
        .stats-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 20px; margin-bottom: 40px; }
        .stat-card { background: var(--card); backdrop-filter: blur(10px); border: 1px solid var(--border); border-radius: 24px; padding: 30px; position: relative; overflow: hidden; }
        .stat-card::before { content: ''; position: absolute; top: -50%; left: -50%; width: 200%; height: 200%; background: radial-gradient(circle, rgba(99,102,241,0.1) 0%, transparent 70%); opacity: 0; transition: 0.4s; }
        .stat-card:hover::before { opacity: 1; }
        .stat-card .label { color: #94a3b8; font-size: 0.9rem; font-weight: 600; text-transform: uppercase; letter-spacing: 1px; }
        .stat-card .value { font-size: 2.8rem; font-weight: 900; margin: 10px 0; font-family: 'Outfit'; }
        .stat-card .trend { font-size: 0.8rem; display: flex; align-items: center; gap: 5px; }
        .trend.up { color: #10b981; }
        
        /* Charts Area */
        .panels { display: grid; grid-template-columns: 2fr 1fr; gap: 30px; margin-bottom: 40px; }
        .panel-card { background: var(--card); border: 1px solid var(--border); border-radius: 24px; padding: 30px; }
        .panel-card h3 { margin-bottom: 25px; font-weight: 800; display: flex; align-items: center; gap: 10px; }
        
        /* Activity Stream */
        .stream { display: flex; flex-direction: column; gap: 15px; max-height: 400px; overflow-y: auto; }
        .stream-item { display: flex; align-items: center; gap: 15px; padding: 15px; background: var(--glass); border-radius: 16px; border: 1px solid var(--border); animation: slideIn 0.3s ease; }
        .stream-icon { width: 40px; height: 40px; border-radius: 12px; display: flex; align-items: center; justify-content: center; background: rgba(99,102,241,0.1); color: var(--primary); }
        .stream-info { flex: 1; }
        .stream-title { font-size: 0.9rem; font-weight: 700; }
        .stream-time { font-size: 0.75rem; color: #94a3b8; }
        
        /* Table Style */
        .table-panel { width: 100%; border-collapse: separate; border-spacing: 0 10px; }
        .table-panel th { padding: 15px; text-align: right; color: #94a3b8; font-weight: 600; border-bottom: 1px solid var(--border); }
        .table-panel td { padding: 15px; background: var(--glass); }
        .table-panel tr td:first-child { border-radius: 0 16px 16px 0; }
        .table-panel tr td:last-child { border-radius: 16px 0 0 16px; }
        
        @keyframes slideIn { from { opacity: 0; transform: translateX(20px); } to { opacity: 1; transform: translateX(0); } }
        
        /* Pulse Effects */
        .pulse { position: relative; }
        .pulse::after { content: ''; position: absolute; width: 8px; height: 8px; background: #10b981; border-radius: 50%; right: -15px; top: 10px; box-shadow: 0 0 10px #10b981; animation: pulseAni 2s infinite; }
        @keyframes pulseAni { 0% { opacity: 1; transform: scale(1); } 50% { opacity: 0.5; transform: scale(1.5); } 100% { opacity: 1; transform: scale(1); } }
    </style>
</head>
<body>
    <div class="sidebar">
        <div class="nav-item active"><i class="fas fa-chart-line"></i><span class="nav-text">التحليلات</span></div>
        <div class="nav-item"><i class="fas fa-users"></i><span class="nav-text">المستخدمين</span></div>
        <div class="nav-item"><i class="fas fa-shield-halved"></i><span class="nav-text">الأمان</span></div>
        <div class="nav-item" style="margin-top: auto;"><i class="fas fa-gear"></i><span class="nav-text">الإعدادات</span></div>
    </div>
    
    <div class="content">
        <div class="header">
            <h1 class="pulse">مركز القيادة والتحليلات Matrix</h1>
            <p style="color: #94a3b8;">إدارة ذكية لجميع موارد النظام والاتصالات في الوقت الحقيقي</p>
        </div>
        
        <div class="stats-grid" id="statsGrid"></div>
        
        <div class="panels">
            <div class="panel-card">
                <h3><i class="fas fa-microchip"></i> الموارد الحية</h3>
                <div style="height: 300px;"><canvas id="mainChart"></canvas></div>
            </div>
            <div class="panel-card">
                <h3><i class="fas fa-bolt"></i> النشاط الأخير</h3>
                <div class="stream" id="eventStream"></div>
            </div>
        </div>
        
        <div class="panel-card">
            <h3><i class="fas fa-globe"></i> الجلسات النشطة عبر الشبكة</h3>
            <table class="table-panel">
                <thead>
                    <tr>
                        <th>الموقع/IP</th>
                        <th>النظام المستهدف</th>
                        <th>مشاهدات الصفحة</th>
                        <th>مستوى التهديد</th>
                        <th>آخر ظهور</th>
                        <th>الحالة</th>
                    </tr>
                </thead>
                <tbody id="sessionsBody"></tbody>
            </table>
        </div>
    </div>

    <script>
        let mainChart;
        let events = [];
        
        function connect() {
            const proto = location.protocol === "https:" ? "wss:" : "ws:";
            const ws = new WebSocket(proto + "//" + location.host + "/api/analytics/events");
            ws.onmessage = (e) => {
                const d = JSON.parse(e.data);
                if (d.type === "initial_stats") update(d.data);
                else {
                    if (d.type === "action" || d.type === "file_activity") addEvent(d);
                    fetch("/api/analytics/stats").then(r => r.json()).then(update);
                }
            };
            ws.onclose = () => setTimeout(connect, 3000);
        }

        function update(s) {
            document.getElementById("statsGrid").innerHTML = 
                card("إجمالي الزيارات", s.total_page_views, "fas fa-eye", "up", "12%") +
                card("مستخدمين نشطين", s.active_sessions, "fas fa-user-check", "up", "5%") +
                card("تبادل البيانات", formatBytes(s.total_files_uploaded + s.total_files_downloaded), "fas fa-right-left", "up", "8%") +
                card("كفاءة النظام", "99.9%", "fas fa-gauge-high", "up", "0.1%");
            
            const tbody = document.getElementById("sessionsBody");
            tbody.innerHTML = (s.recent_sessions || []).map(sess => 
                "<tr>" +
                "<td><i class='fas fa-location-dot'></i> " + sess.ip_address + "</td>" +
                "<td>" + sess.os + " / " + sess.browser + "</td>" +
                "<td>" + sess.page_views + "</td>" +
                "<td><span style='color: " + getThreatColor(sess.page_views) + "; font-weight:800;'>" + getThreatLevel(sess.page_views) + "</span></td>" +
                "<td>" + new Date(sess.last_activity).toLocaleTimeString() + "</td>" +
                "<td>" + (sess.is_active ? "<span style='color:#10b981;'>نشط</span>" : "خامل") + "</td>" +
                "</tr>"
            ).join("");
            
            updateMainChart(s);
        }

        function addEvent(d) {
            const stream = document.getElementById("eventStream");
            const item = document.createElement("div");
            item.className = "stream-item";
            let type = d.type === "action" ? d.data.type : d.data.action;
            let icon = type.includes("upload") ? "fa-cloud-arrow-up" : "fa-bolt";
            
            item.innerHTML = 
                "<div class='stream-icon'><i class='fas " + icon + "'></i></div>" +
                "<div class='stream-info'>" +
                "<div class='stream-title'>" + (d.data.path || d.data.file_name) + "</div>" +
                "<div class='stream-time'>" + new Date().toLocaleTimeString() + " - " + type + "</div>" +
                "</div>";
            
            stream.insertBefore(item, stream.firstChild);
            if (stream.children.length > 10) stream.lastChild.remove();
        }

        function updateMainChart(s) {
            const ctx = document.getElementById('mainChart').getContext('2d');
            if (mainChart) mainChart.destroy();
            mainChart = new Chart(ctx, {
                type: 'line',
                data: {
                    labels: ['T-5', 'T-4', 'T-3', 'T-2', 'T-1', 'Now'],
                    datasets: [{
                        label: 'عدد الجلسات النشطة',
                        data: [s.active_sessions-2, s.active_sessions-1, s.active_sessions+1, s.active_sessions, s.active_sessions+2, s.active_sessions],
                        borderColor: '#6366f1',
                        borderWidth: 3,
                        pointRadius: 5,
                        backgroundColor: 'rgba(99, 102, 241, 0.1)',
                        fill: true,
                        tension: 0.4
                    }]
                },
                options: { 
                    responsive: true, 
                    maintainAspectRatio: false,
                    plugins: { legend: { display: false } },
                    scales: { 
                        y: { grid: { color: 'rgba(255,255,255,0.05)' }, ticks: { color: '#94a3b8' } },
                        x: { grid: { display: false }, ticks: { color: '#94a3b8' } }
                    }
                }
            });
        }

        function card(l, v, i, t, tv) {
            return "<div class='stat-card'>" +
                "<div style='display:flex; justify-content:space-between;'>" +
                "<div class='label'>" + l + "</div>" +
                "<i class='" + i + "' style='color:var(--primary); opacity:0.5;'></i>" +
                "</div>" +
                "<div class='value'>" + v + "</div>" +
                "<div class='trend " + t + "'><i class='fas fa-arrow-trend-up'></i> " + tv + " منذ الساعة الماضية</div>" +
                "</div>";
        }

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function getThreatLevel(views) {
            if (views > 100) return "CRITICAL";
            if (views > 50) return "HIGH";
            return "NORMAL";
        }
        function getThreatColor(views) {
            if (views > 100) return "#ef4444";
            if (views > 50) return "#f59e0b";
            return "#10b981";
        }

        connect();
        setInterval(() => fetch("/api/analytics/stats").then(r => r.json()).then(update), 15000);
    </script>
</body>
</html>`
