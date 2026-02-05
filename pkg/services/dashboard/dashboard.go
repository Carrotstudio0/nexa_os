package dashboard

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/utils"
)

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	utils.LogInfo("Dashboard", "Connection received from: "+r.RemoteAddr)
	localIP := utils.GetLocalIP()

	data := map[string]interface{}{
		"LocalIP":  localIP,
		"Port":     config.DashboardPort,
		"Services": config.Services,
	}

	tmpl, err := template.New("dashboard").Parse(dashboardHTML)
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), 500)
		return
	}
	tmpl.Execute(w, data)
}

func handleProxyFiles(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/storage")
	if path == "" {
		path = "/"
	}
	target := "http://127.0.0.1:" + config.WebPort + path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	req, _ := http.NewRequest(r.Method, target, r.Body)
	for k, v := range r.Header {
		req.Header[k] = v
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Service Unavailable", 503)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleProxyAdmin(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin")
	if path == "" {
		path = "/"
	}
	target := "http://127.0.0.1:" + config.AdminPort + path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	req, _ := http.NewRequest(r.Method, target, r.Body)
	for k, v := range r.Header {
		req.Header[k] = v
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Service Unavailable", 503)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func handleProxyChat(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/chat")
	if path == "" {
		path = "/"
	}
	target := "http://127.0.0.1:" + config.ChatPort + path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	req, _ := http.NewRequest(r.Method, target, r.Body)
	for k, v := range r.Header {
		req.Header[k] = v
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Service Unavailable", 503)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDashboard)
	mux.HandleFunc("/storage/", handleProxyFiles)
	mux.HandleFunc("/admin/", handleProxyAdmin)
	mux.HandleFunc("/chat/", handleProxyChat)

	localIP := utils.GetLocalIP()
	utils.LogInfo("Dashboard", fmt.Sprintf("Web Interface:     http://%s:%s", localIP, config.DashboardPort))
	utils.SaveEndpoint("dashboard", fmt.Sprintf("http://%s:%s", localIP, config.DashboardPort))

	server := &http.Server{
		Addr:    ":" + config.DashboardPort,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.LogFatal("Dashboard", err.Error())
	}
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NEXA ULTIMATE | Command Center</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;800&family=Cairo:wght@400;600;700;900&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/axios/dist/axios.min.js"></script>
    <style>
        :root {
            --primary: #6366f1;
            --secondary: #ec4899;
            --accent: #06b6d4;
            --bg: #020617;
            --card-bg: rgba(15, 23, 42, 0.6);
            --glass: rgba(255, 255, 255, 0.03);
            --border: rgba(255, 255, 255, 0.08);
            --text: #f8fafc;
            --text-muted: #64748b;
            --sidebar-width: 280px;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        body {
            font-family: 'Outfit', 'Cairo', sans-serif;
            background: var(--bg);
            background-image: 
                radial-gradient(circle at 0% 0%, rgba(99, 102, 241, 0.1) 0%, transparent 40%),
                radial-gradient(circle at 100% 100%, rgba(236, 72, 153, 0.1) 0%, transparent 40%);
            color: var(--text);
            height: 100vh;
            overflow: hidden;
            display: flex;
        }

        /* Sidebar Glassmorphism */
        .sidebar {
            width: var(--sidebar-width);
            background: rgba(2, 6, 23, 0.8);
            backdrop-filter: blur(20px);
            -webkit-backdrop-filter: blur(20px);
            border-left: 1px solid var(--border);
            display: flex;
            flex-direction: column;
            padding: 40px 24px;
            z-index: 100;
        }

        .logo {
            display: flex;
            align-items: center;
            gap: 16px;
            font-size: 1.75rem;
            font-weight: 900;
            background: linear-gradient(to right, var(--primary), var(--secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 60px;
            letter-spacing: -1px;
        }

        .nav-group { margin-bottom: 40px; }
        .nav-label {
            font-size: 0.75rem;
            text-transform: uppercase;
            letter-spacing: 2px;
            color: var(--text-muted);
            margin-bottom: 16px;
            padding-right: 12px;
        }

        .nav-item {
            padding: 14px 16px;
            margin-bottom: 8px;
            border-radius: 16px;
            color: var(--text-muted);
            cursor: pointer;
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            display: flex;
            align-items: center;
            gap: 16px;
            font-weight: 600;
            border: 1px solid transparent;
        }

        .nav-item:hover {
            background: var(--glass);
            color: var(--text);
            transform: translateX(-4px);
        }

        .nav-item.active {
            background: linear-gradient(135deg, rgba(99, 102, 241, 0.15), rgba(236, 72, 153, 0.15));
            color: var(--text);
            border-color: rgba(99, 102, 241, 0.3);
            box-shadow: 0 10px 20px -10px rgba(0, 0, 0, 0.5);
        }

        .nav-item i { font-size: 1.25rem; width: 24px; }

        .sys-info {
            margin-top: auto;
            background: var(--glass);
            border: 1px solid var(--border);
            border-radius: 20px;
            padding: 20px;
        }

        .status-pill {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            background: rgba(34, 197, 94, 0.1);
            color: #4ade80;
            padding: 4px 12px;
            border-radius: 100px;
            font-size: 0.8rem;
            font-weight: 700;
        }

        .pulse {
            width: 8px; height: 8px; background: #4ade80; border-radius: 50%;
            animation: pulse-ring 1.5s infinite;
        }

        @keyframes pulse-ring {
            0% { transform: scale(0.8); box-shadow: 0 0 0 0 rgba(74, 222, 128, 0.7); }
            70% { transform: scale(1); box-shadow: 0 0 0 10px rgba(74, 222, 128, 0); }
            100% { transform: scale(0.8); box-shadow: 0 0 0 0 rgba(74, 222, 128, 0); }
        }

        /* Main Content */
        .main {
            flex: 1;
            padding: 60px;
            overflow-y: auto;
            background: transparent;
        }

        .section { display: none; animation: slideUp 0.6s cubic-bezier(0.16, 1, 0.3, 1); }
        .section.active { display: block; }

        @keyframes slideUp {
            from { opacity: 0; transform: translateY(30px); }
            to { opacity: 1; transform: translateY(0); }
        }

        header h1 { font-size: 3rem; font-weight: 900; margin-bottom: 40px; }

        /* Metrics */
        .metrics {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
            gap: 24px;
            margin-bottom: 48px;
        }

        .metric-card {
            background: var(--card-bg);
            border: 1px solid var(--border);
            border-radius: 28px;
            padding: 32px;
            position: relative;
            overflow: hidden;
            transition: all 0.3s;
        }

        .metric-card:hover { border-color: var(--primary); transform: translateY(-4px); }

        .m-label { font-size: 0.85rem; text-transform: uppercase; color: var(--text-muted); letter-spacing: 1.5px; margin-bottom: 12px; }
        .m-value { font-size: 2.5rem; font-weight: 800; }
        .m-icon { position: absolute; right: -20px; bottom: -20px; font-size: 6rem; opacity: 0.05; transform: rotate(-15deg); }

        /* Services Grid */
        .service-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
            gap: 24px;
        }

        .s-card {
            background: var(--glass);
            border: 1px solid var(--border);
            border-radius: 32px;
            padding: 40px;
            display: flex;
            flex-direction: column;
            gap: 24px;
            transition: all 0.4s;
        }

        .s-card:hover {
            background: rgba(255, 255, 255, 0.06);
            border-color: var(--secondary);
            transform: scale(1.02);
            box-shadow: 0 40px 80px -20px rgba(0, 0, 0, 0.6);
        }

        .s-header { display: flex; align-items: center; gap: 20px; }
        .s-icon { 
            width: 64px; height: 64px; background: rgba(255, 255, 255, 0.05); 
            border-radius: 18px; display: flex; align-items: center; justify-content: center;
            font-size: 1.75rem; border: 1px solid var(--border);
        }

        .s-info h3 { font-size: 1.5rem; font-weight: 800; }
        .s-info p { color: var(--text-muted); font-size: 1rem; line-height: 1.6; }

        .btn-glow {
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            color: white; border: none; padding: 16px 32px;
            border-radius: 16px; cursor: pointer; font-weight: 700;
            text-decoration: none; text-align: center;
            transition: all 0.3s;
        }

        .btn-glow:hover {
            box-shadow: 0 0 30px rgba(99, 102, 241, 0.4);
            transform: translateY(-2px);
        }

        /* Responsive */
        @media (max-width: 1024px) {
            .sidebar { width: 80px; padding: 40px 10px; }
            .logo span, .nav-label, .nav-item span, .sys-info { display: none; }
            .nav-item { justify-content: center; }
            .main { padding: 30px; }
        }
    </style>
</head>
<body>
    <aside class="sidebar">
        <div class="logo">
            <i class="fas fa-microchip"></i>
            <span>NEXA OS</span>
        </div>

        <div class="nav-group">
            <div class="nav-label">التحكم</div>
            <div class="nav-item active" onclick="showSection('overview', this)">
                <i class="fas fa-shapes"></i>
                <span>الرئيسية</span>
            </div>
            <div class="nav-item" onclick="showSection('files', this)">
                <i class="fas fa-database"></i>
                <span>المخزن الرقمي</span>
            </div>
            <div class="nav-item" onclick="showSection('chat', this)">
                <i class="fas fa-meteor"></i>
                <span>الاتصال الكمي</span>
            </div>
        </div>

        <div class="nav-group">
            <div class="nav-label">النظام</div>
            <div class="nav-item" onclick="showSection('admin', this)">
                <i class="fas fa-gears"></i>
                <span>الإدارة</span>
            </div>
            <div class="nav-item" onclick="showSection('network', this)">
                <i class="fas fa-satellite-dish"></i>
                <span>خريطة الشبكة</span>
            </div>
        </div>

        <div class="sys-info">
            <div class="status-pill">
                <div class="pulse"></div>
                النظام مستقر
            </div>
            <div style="margin-top:16px; color:var(--text-muted); font-size:0.85rem;">
                IP: <span style="color:var(--text); font-weight:700;">{{.LocalIP}}</span>
            </div>
        </div>
    </aside>

    <main class="main">
        <div id="overview" class="section active">
            <header>
                <h1>Command Center</h1>
            </header>

            <div class="metrics">
                <div class="metric-card">
                    <div class="m-label">العمليات النشطة</div>
                    <div class="m-value">12.4k/s</div>
                    <i class="fas fa-bolt m-icon"></i>
                </div>
                <div class="metric-card">
                    <div class="m-label">الأجهزة المتصلة</div>
                    <div class="m-value" id="device-count">3</div>
                    <i class="fas fa-link m-icon"></i>
                </div>
                <div class="metric-card" style="background: linear-gradient(135deg, rgba(99,102,241,0.1), transparent);">
                    <div class="m-label">زمن الاستجابة</div>
                    <div class="m-value">0.4ms</div>
                    <i class="fas fa-stopwatch m-icon"></i>
                </div>
            </div>

            <div class="service-grid">
                <div class="s-card" style="cursor:pointer;" onclick="showSection('files', document.querySelectorAll('.nav-item')[1])">
                    <div class="s-header">
                        <div class="s-icon">📁</div>
                        <div class="s-info">
                            <h3>المخزن الرقمي</h3>
                            <p>Port 8081 | نقل وإدارة الملفات</p>
                        </div>
                    </div>
                    <button class="btn-glow">دخول الوحدة ←</button>
                </div>
                <div class="s-card" style="cursor:pointer;" onclick="showSection('chat', document.querySelectorAll('.nav-item')[2])">
                    <div class="s-header">
                        <div class="s-icon">💬</div>
                        <div class="s-info">
                            <h3>الاتصال الكمي</h3>
                            <p>Port 8082 | محادثة مشفرة فورية</p>
                        </div>
                    </div>
                    <button class="btn-glow">دخول الوحدة ←</button>
                </div>
                <div class="s-card" style="cursor:pointer;" onclick="showSection('admin', document.querySelectorAll('.nav-item')[4])">
                    <div class="s-header">
                        <div class="s-icon">⚙️</div>
                        <div class="s-info">
                            <h3>إدارة النظام</h3>
                            <p>Port 8080 | التحكم والإعدادات</p>
                        </div>
                    </div>
                    <button class="btn-glow">دخول الوحدة ←</button>
                </div>
            </div>
        </div>

        <!-- File Storage Section -->
        <div id="files" class="section">
            <header>
                <h1>📁 المخزن الرقمي</h1>
                <a href="http://{{.LocalIP}}:8081/" target="_blank" class="btn-glow" style="font-size: 0.8rem; padding: 10px 20px;">فتح في نافذة جديدة ↗</a>
            </header>
            <iframe src="/storage/" style="width:100%; height:calc(100vh - 200px); border:none; border-radius:32px; background:var(--bg);"></iframe>
        </div>

        <!-- Chat Section -->
        <div id="chat" class="section">
            <header>
                <h1>💬 الاتصال الكمي</h1>
                <a href="http://{{.LocalIP}}:8082/" target="_blank" class="btn-glow" style="font-size: 0.8rem; padding: 10px 20px;">فتح في نافذة جديدة ↗</a>
            </header>
            <iframe src="/chat/" style="width:100%; height:calc(100vh - 200px); border:none; border-radius:32px; background:var(--bg);"></iframe>
        </div>

        <!-- Admin Section -->
        <div id="admin" class="section">
            <header>
                <h1>⚙️ إدارة النظام</h1>
                <a href="http://{{.LocalIP}}:8080/" target="_blank" class="btn-glow" style="font-size: 0.8rem; padding: 10px 20px;">فتح في نافذة جديدة ↗</a>
            </header>
            <iframe src="/admin/" style="width:100%; height:calc(100vh - 200px); border:none; border-radius:32px; background:var(--bg);"></iframe>
        </div>

        <!-- Network Map Section -->
        <div id="network" class="section">
            <header><h1>🌐 خريطة الشبكة</h1></header>
            <div style="background:var(--card-bg); border:1px solid var(--border); border-radius:32px; padding:40px; height:calc(100vh - 200px); overflow-y:auto;">
                <div style="text-align:center; padding:60px; color:var(--text-muted);">
                    <i class="fas fa-network-wired" style="font-size:4rem; margin-bottom:20px; opacity:0.3;"></i>
                    <h2>خريطة الشبكة النشطة</h2>
                    <p style="margin-top:10px;">جميع الأجهزة المتصلة والخوادم النشطة</p>
                    <div style="margin-top:40px; background:var(--glass); border:1px solid var(--border); border-radius:20px; padding:30px;">
                        <h3>الأجهزة المتصلة: 1</h3>
                        <p style="margin-top:10px;">Server IP: {{.LocalIP}}</p>
                        <p>Gateway: TCP:8000</p>
                        <p>Dashboard: TCP:7000</p>
                    </div>
                </div>
            </div>
        </div>
    </main>

    <script>
        function showSection(id, el) {
            document.querySelectorAll('.section').forEach(s => s.classList.remove('active'));
            document.querySelectorAll('.nav-item').forEach(i => i.classList.remove('active'));
            document.getElementById(id).classList.add('active');
            el.classList.add('active');
        }
    </script>
</body>
</html>
`
