package main

import (
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
)

const DashboardPort = "7000"

// Services Configuration
var services = []struct {
	Name, Port, Description, Icon, URL string
}{
	{"Gateway", "8000", "البوابة الرئيسية", "🌐", "http://{{.LocalIP}}:8000"},
	{"Admin Panel", "8080", "لوحة التحكم", "⚙️", "http://{{.LocalIP}}:8080"},
	{"File Manager", "8081", "مدير الملفات", "💾", "http://{{.LocalIP}}:8081"},
	{"DNS Server", "53", "خادم DNS", "🔍", "dns://{{.LocalIP}}:53"},
	{"Core Server", "9000", "الخادم الأساسي", "🖥️", "http://{{.LocalIP}}:9000"},
}

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

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	log.Printf("📥 Connection received from: %s", r.RemoteAddr)
	localIP := getLocalIP()
	data := map[string]interface{}{
		"LocalIP":  localIP,
		"Port":     DashboardPort,
		"Services": services,
	}

	tmpl, err := template.New("dashboard").Parse(dashboardHTML)
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), 500)
		return
	}
	tmpl.Execute(w, data)
}

func main() {
	localIP := getLocalIP()
	fmt.Println(`
     _   _  ________  _______     _______                    
    | \ | |/ ____\  \/  / __ \   |  __ \                   
    |  \| | |  __ \   _/ |  | |  | |  | |_ __ ___  _ __    
    | .   | |_| |  >  <| |  | |  | |  | | '__/ _ \| '_ \   
    | |\  | |__| |/  . \ |__| |  | |__| | | | (_) | |_) |  
    |_| \_|\_____/_/ \_\____/   |_____/|_|  \___/| .__/   
                                                 | |
                                                 |_|  v3.0 Ultimate`)
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Printf("   [INFO]  Initializing Dashboard UI...\n")
	fmt.Printf("   [INFO]  Serving UI at:     http://%s:%s\n", localIP, DashboardPort)
	fmt.Printf("   [INFO]  Access-Log:        %s\n", "ENABLED")
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Println("   ✅  DASHBOARD ONLINE")

	http.HandleFunc("/", handleDashboard)

	if err := http.ListenAndServe(":"+DashboardPort, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

const dashboardHTML = `
<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Nexa OS | نظام التحكم الشامل</title>
    <link href="https://fonts.googleapis.com/css2?family=Cairo:wght@300;400;600;700;800&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/axios/dist/axios.min.js"></script>
    <style>
        :root {
            --bg-dark: #0f0c29;
            --bg-gradient: linear-gradient(135deg, #0f0c29, #302b63, #24243e);
            --glass: rgba(255, 255, 255, 0.05);
            --glass-border: rgba(255, 255, 255, 0.1);
            --primary: #00d2ff;
            --secondary: #3a7bd5;
            --text: #ffffff;
            --text-muted: #b0b0b0;
            --sidebar-width: 280px;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font-family: 'Cairo', sans-serif;
            background: var(--bg-gradient);
            color: var(--text);
            height: 100vh;
            overflow: hidden;
            display: flex;
        }

        /* Sidebar */
        .sidebar {
            width: var(--sidebar-width);
            background: rgba(0,0,0,0.3);
            backdrop-filter: blur(20px);
            border-left: 1px solid var(--glass-border);
            display: flex;
            flex-direction: column;
            padding: 30px 20px;
            z-index: 100;
        }

        .logo {
            font-size: 2rem;
            font-weight: 800;
            background: linear-gradient(to right, #00d2ff, #3a7bd5);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 50px;
            display: flex;
            align-items: center;
            gap: 15px;
        }

        .nav-item {
            padding: 15px 20px;
            margin-bottom: 10px;
            border-radius: 15px;
            color: var(--text-muted);
            cursor: pointer;
            transition: all 0.3s ease;
            display: flex;
            align-items: center;
            gap: 15px;
            font-weight: 600;
        }

        .nav-item:hover, .nav-item.active {
            background: linear-gradient(90deg, rgba(0, 210, 255, 0.1), transparent);
            color: var(--primary);
            border-right: 3px solid var(--primary);
        }

        .nav-item i { font-size: 1.2rem; min-width: 25px; }

        .system-status {
            margin-top: auto;
            padding: 20px;
            background: rgba(0,0,0,0.2);
            border-radius: 15px;
        }

        .status-dot {
            width: 10px; height: 10px; background: #00ff88;
            border-radius: 50%; display: inline-block;
            box-shadow: 0 0 10px #00ff88; margin-left: 10px;
        }

        /* Main Content */
        .main-content {
            flex: 1;
            padding: 40px;
            overflow-y: auto;
            position: relative;
        }

        .section { display: none; animation: fadeIn 0.5s ease; }
        .section.active { display: block; }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }

        h1 { font-size: 2.5rem; margin-bottom: 30px; font-weight: 700; }

        /* Dashboard Cards */
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 25px;
        }

        .card {
            background: var(--glass);
            border: 1px solid var(--glass-border);
            border-radius: 20px;
            padding: 25px;
            transition: transform 0.3s;
            position: relative;
            overflow: hidden;
        }

        .card:hover { transform: translateY(-5px); background: rgba(255,255,255,0.08); }

        .card::before {
            content: ''; position: absolute; top:0; left:0; right:0; height: 4px;
            background: linear-gradient(90deg, var(--primary), var(--secondary));
        }

        .card-title { color: var(--text-muted); font-size: 0.9rem; margin-bottom: 15px; }
        .card-value { font-size: 2rem; font-weight: 700; margin-bottom: 5px; }
        .card-icon { 
            position: absolute; left: 20px; bottom: 20px; 
            font-size: 3rem; opacity: 0.1; 
        }

        /* Services List */
        .service-list {
            margin-top: 40px;
            background: var(--glass);
            border-radius: 20px;
            overflow: hidden;
            border: 1px solid var(--glass-border);
        }

        .service-item {
            display: flex;
            align-items: center;
            padding: 20px 30px;
            border-bottom: 1px solid var(--glass-border);
            transition: background 0.3s;
        }

        .service-item:last-child { border-bottom: none; }
        .service-item:hover { background: rgba(255,255,255,0.03); }

        .service-icon { 
            font-size: 1.5rem; width: 50px; height: 50px; 
            background: rgba(255,255,255,0.1); border-radius: 12px;
            display: flex; align-items: center; justify-content: center;
            margin-left: 20px;
        }

        .service-info h3 { font-size: 1.1rem; margin-bottom: 5px; }
        .service-info p { color: var(--text-muted); font-size: 0.9rem; }
        
        .service-action { margin-right: auto; }
        .btn {
            background: linear-gradient(90deg, var(--primary), var(--secondary));
            color: white; border: none; padding: 10px 25px;
            border-radius: 10px; cursor: pointer; font-weight: 600;
            font-family: inherit; text-decoration: none; display: inline-block;
            transition: shadow 0.3s;
        }
        .btn:hover { box-shadow: 0 5px 20px rgba(0, 210, 255, 0.3); }

        /* File Manager */
        .file-manager {
            background: var(--glass);
            border-radius: 20px;
            border: 1px solid var(--glass-border);
            height: calc(100vh - 120px);
            display: flex;
            flex-direction: column;
        }

        .fm-toolbar {
            padding: 20px;
            border-bottom: 1px solid var(--glass-border);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .fm-grid {
            padding: 20px;
            overflow-y: auto;
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
            gap: 20px;
        }

        .file-item {
            background: rgba(255,255,255,0.05);
            border-radius: 15px;
            padding: 15px;
            text-align: center;
            cursor: pointer;
            transition: all 0.2s;
        }
        .file-item:hover { background: rgba(255,255,255,0.1); transform: scale(1.05); }
        .file-icon-lg { font-size: 3rem; margin-bottom: 10px; display: block; }
        .file-name { font-size: 0.9rem; word-break: break-word; color: #ddd; }
        .file-meta { font-size: 0.75rem; color: #888; margin-top: 5px; }

        /* Iframe for Admin */
        .full-frame {
            width: 100%;
            height: calc(100vh - 100px);
            border: none;
            border-radius: 20px;
            background: white;
        }

        .upload-zone {
            border: 2px dashed var(--glass-border);
            border-radius: 15px;
            padding: 30px;
            text-align: center;
            margin-bottom: 20px;
            transition: all 0.3s;
            cursor: pointer;
        }
        .upload-zone:hover { border-color: var(--primary); background: rgba(0,210,255,0.05); }

        /* Loader */
        .loader { text-align: center; padding: 50px; color: var(--text-muted); }
    </style>
</head>
<body>
    <!-- Sidebar -->
    <div class="sidebar">
        <div class="logo">
            <i class="fas fa-cube"></i> Nexa OS
        </div>
        
        <div class="nav-item active" onclick="showSection('overview', this)">
            <i class="fas fa-home"></i> نظرة عامة
        </div>
        <div class="nav-item" onclick="showSection('files', this)">
            <i class="fas fa-folder-open"></i> الملفات
        </div>
        <div class="nav-item" onclick="showSection('chat', this)">
            <i class="fas fa-comments"></i> المحادثة
        </div>
        <div class="nav-item" onclick="showSection('admin', this)">
            <i class="fas fa-cogs"></i> إدارة النظام
        </div>
        <div class="nav-item" onclick="showSection('network', this)">
            <i class="fas fa-network-wired"></i> الشبكة
        </div>

        <div class="system-status">
            <div style="margin-bottom:5px; font-weight:bold;">حالة النظام</div>
            <div style="font-size:0.85rem; color:#aaa;">
                <span class="status-dot"></span> متصل
                <br> IP: {{.LocalIP}}
            </div>
        </div>
        <div style="margin-top:20px; font-size:0.8rem; color:#666; text-align:center;">
            v3.0.0 Ultimate
        </div>
    </div>

    <!-- Main Content -->
    <div class="main-content">
        
        <!-- OVERVIEW SECTION -->
        <div id="overview" class="section active">
            <h1>لوحة القيادة المركزية</h1>
            
            <div class="grid">
                <div class="card">
                    <div class="card-title">الخدمات النشطة</div>
                    <div class="card-value" id="active-services">5</div>
                    <div class="card-icon">⚡</div>
                    <div style="color:#00ff88; font-size:0.9rem;">▲ النظام يعمل بكفاءة</div>
                </div>
                <div class="card">
                    <div class="card-title">الملفات المخزنة</div>
                    <div class="card-value" id="total-files">--</div>
                    <div class="card-icon">📂</div>
                </div>
                <div class="card">
                    <div class="card-title">حجم البيانات</div>
                    <div class="card-value" id="total-size">--</div>
                    <div class="card-icon">💾</div>
                </div>
            </div>

            <div class="service-list">
                {{range .Services}}
                <div class="service-item">
                    <div class="service-icon">{{.Icon}}</div>
                    <div class="service-info">
                        <h3>{{.Name}}</h3>
                        <p>{{.Description}}</p>
                    </div>
                    <div class="service-action">
                        <span style="background:rgba(0,255,136,0.1); color:#00ff88; padding:5px 10px; border-radius:8px; margin-left:10px; font-size:0.9rem;">● نشط</span>
                        <a href="{{.URL}}" target="_blank" class="btn" style="padding:8px 20px; font-size:0.9rem;">فتح</a>
                    </div>
                </div>
                {{end}}
            </div>
        </div>

        <!-- FILES SECTION -->
        <div id="files" class="section">
            <h1>مدير الملفات المتكامل</h1>
            <div class="file-manager">
                <div class="fm-toolbar">
                    <div style="display:flex; gap:10px;">
                        <input type="file" id="fileUploadInput" hidden multiple>
                        <button class="btn" onclick="document.getElementById('fileUploadInput').click()">
                            <i class="fas fa-upload"></i> رفع ملفات
                        </button>
                        <button class="btn" onclick="loadFiles()" style="background:#444;">
                            <i class="fas fa-sync"></i> تحديث
                        </button>
                    </div>
                    <div>
                        <span id="files-status" style="color:var(--text-muted)">جاري التحميل...</span>
                    </div>
                </div>
                
                <div id="files-container" class="fm-grid">
                    <!-- Files will be loaded here -->
                    <div class="loader">جاري الاتصال بخادم الملفات...</div>
                </div>
            </div>
        </div>

        <!-- ADMIN SECTION -->
        <div id="admin" class="section">
            <h1 style="margin-bottom:10px;">وحدة التحكم الإدارية</h1>
            <p style="color:var(--text-muted); margin-bottom:20px;">وصول مباشر لنظام الإدارة (Admin Panel)</p>
            <iframe src="http://localhost:8080" class="full-frame" title="Admin Panel"></iframe>
        </div>

        <!-- CHAT SECTION -->
        <div id="chat" class="section">
            <h1>المحادثة العامة 💬</h1>
            <div class="chat-container">
                <div class="chat-messages" id="chat-messages">
                    <!-- Messages will appear here -->
                    <div style="text-align:center; color:#666; margin-top:50px;">جاري تحميل المحادثة...</div>
                </div>
                <div class="chat-input-area">
                    <input type="text" id="chat-username" placeholder="الاسم" style="width:20%; padding:15px; border-radius:10px; border:none; background:rgba(255,255,255,0.1); color:white; text-align:center;">
                    <input type="text" id="chat-input" placeholder="اكتب رسالتك هنا..." style="width:60%; padding:15px; border-radius:10px; border:none; background:rgba(255,255,255,0.1); color:white;">
                    <button class="btn" onclick="sendMessage()" style="width:15%;">إرسال 🚀</button>
                </div>
            </div>
        </div>

        <!-- NETWORK SECTION (Placeholder) -->
        <div id="network" class="section">
            <h1>مراقبة الشبكة</h1>
            <div class="card" style="height:400px; display:flex; align-items:center; justify-content:center; flex-direction:column;">
                <i class="fas fa-globe-americas" style="font-size:5rem; color:var(--primary); margin-bottom:20px; opacity:0.5;"></i>
                <h2>خريطة الشبكة الحية</h2>
                <p style="color:var(--text-muted)">جاري مسح الأجهزة المتصلة...</p>
                <div style="margin-top:20px; font-family:monospace; color:#00ff88;">
                    > Scanning 192.168.1.0/24...<br>
                    > Localhost found (127.0.0.1)<br>
                    > Gateway active
                </div>
            </div>
        </div>

    </div>

    <style>
        /* Chat Styles */
        .chat-container {
            background: var(--glass);
            border-radius: 20px;
            border: 1px solid var(--glass-border);
            height: calc(100vh - 120px);
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        .chat-messages {
            flex: 1;
            padding: 20px;
            overflow-y: auto;
            display: flex;
            flex-direction: column;
            gap: 15px;
        }
        .message {
            background: rgba(255,255,255,0.05);
            padding: 10px 15px;
            border-radius: 15px;
            border-bottom-right-radius: 2px;
            max-width: 80%;
            align-self: flex-start;
            animation: fadeIn 0.3s;
        }
        .message.mine {
            background: rgba(0, 210, 255, 0.15);
            align-self: flex-end;
            border-bottom-right-radius: 15px;
            border-bottom-left-radius: 2px;
        }
        .message.admin {
            background: rgba(255, 215, 0, 0.1);
            border: 1px solid rgba(255, 215, 0, 0.3);
            width: 100%;
            text-align: center;
        }
        .msg-header {
            font-size: 0.75rem;
            color: var(--primary);
            margin-bottom: 5px;
            display: flex;
            justify-content: space-between;
        }
        .msg-content { font-size: 1rem; line-height: 1.4; word-break: break-word; }
        .chat-input-area {
            padding: 20px;
            background: rgba(0,0,0,0.2);
            display: flex;
            gap: 10px;
        }
    </style>

    <script>
        // Navigation Logic
        function showSection(id, tab) {
            document.querySelectorAll('.section').forEach(el => el.classList.remove('active'));
            document.getElementById(id).classList.add('active');
            
            if(tab) {
                document.querySelectorAll('.nav-item').forEach(el => el.classList.remove('active'));
                tab.classList.add('active');
            }

            if(id === 'files') loadFiles();
            if(id === 'chat') {
                scrollToBottom();
                document.getElementById('chat-input').focus();
            }
        }

        // Configuration
        const HOST = window.location.hostname;
        const FILES_API = 'http://' + HOST + ':8081';
        const CHAT_API = 'http://' + HOST + ':8082';

        // --- File Manager Logic ---
        async function loadFiles() {
            const container = document.getElementById('files-container');
            const statusParams = document.getElementById('files-status');
            
            try {
                const response = await axios.get(FILES_API + '/api/list');
                const files = response.data;
                
                document.getElementById('total-files').textContent = files.length;
                
                container.innerHTML = '';
                if(files.length === 0) {
                    container.innerHTML = '<div style="grid-column:1/-1; text-align:center; padding:50px; color:#666;">لا توجد ملفات</div>';
                    return;
                }

                files.forEach(file => {
                    const div = document.createElement('div');
                    div.className = 'file-item';
                    div.innerHTML = '<div class="file-icon-lg">' + file.Icon + '</div>' +
                                    '<div class="file-name">' + file.Name + '</div>' +
                                    '<div class="file-meta">' + file.Size + '</div>';
                    div.onclick = () => window.open(FILES_API + '/download?file=' + encodeURIComponent(file.Name));
                    container.appendChild(div);
                });
                statusParams.textContent =  files.length + ' ملفات';
            } catch (error) {
                container.innerHTML = '<div style="color:red; text-align:center; grid-column:1/-1;">خطأ في الاتصال بخادم الملفات (Port 8081)</div>';
            }
        }

        // Upload Logic
        const fileInput = document.getElementById('fileUploadInput');
        fileInput.addEventListener('change', async (e) => {
            if(e.target.files.length === 0) return;
            const formData = new FormData();
            for(let i=0; i<e.target.files.length; i++) {
                formData.append('file', e.target.files[i]);
            }
            try {
                document.getElementById('files-status').textContent = 'جاري الرفع...';
                await axios.post(FILES_API + '/upload', formData, {
                    headers: { 'Content-Type': 'multipart/form-data' }
                });
                loadFiles();
                alert('تم رفع الملفات بنجاح ✅');
            } catch (error) {
                alert('فشل الرفع: ' + error.message);
            }
        });

        // --- Chat Logic ---
        let lastMsgId = 0;
        let myName = localStorage.getItem('nexa_username') || 'Guest';
        document.getElementById('chat-username').value = myName;

        document.getElementById('chat-username').addEventListener('change', (e) => {
            myName = e.target.value;
            localStorage.setItem('nexa_username', myName);
        });

        document.getElementById('chat-input').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') sendMessage();
        });

        async function sendMessage() {
            const input = document.getElementById('chat-input');
            const content = input.value.trim();
            if (!content) return;

            try {
                await axios.post(CHAT_API + '/send', {
                    sender: myName,
                    content: content
                });
                input.value = '';
                loadMessages();
            } catch (error) {
                console.error("Chat Error:", error);
            }
        }

        async function loadMessages() {
            try {
                const response = await axios.get(CHAT_API + '/messages');
                const messages = response.data;
                const container = document.getElementById('chat-messages');

                // Simple render: clear and redraw if count changes (basic syncing)
                // For a real app, we would append only new ones.
                 // Optimization: Only update if length changed or first load
                 // For now, just redraw to be safe and simple.
                container.innerHTML = '';
                
                messages.forEach(msg => {
                    const div = document.createElement('div');
                    const isMine = msg.sender === myName;
                    div.className = 'message ' + (isMine ? 'mine' : '') + (msg.isAdmin ? ' admin' : '');
                    
                    div.innerHTML = '<div class="msg-header">' +
                                        '<span>' + msg.sender + '</span>' +
                                        '<span>' + msg.timestamp + '</span>' +
                                    '</div>' +
                                    '<div class="msg-content">' + msg.content + '</div>';
                    container.appendChild(div);
                });
                
                // Auto scroll if near bottom
                // container.scrollTop = container.scrollHeight;
            } catch (error) {
                console.error("Chat Poll Error:", error);
            }
        }

        function scrollToBottom() {
            const container = document.getElementById('chat-messages');
            container.scrollTop = container.scrollHeight;
        }

        // Initialize
        loadFiles();
        setInterval(loadMessages, 2000); // Poll chat every 2s
        setTimeout(scrollToBottom, 2500);

    </script>
</body>
</html>
`
