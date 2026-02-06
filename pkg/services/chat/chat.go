package chat

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/utils"
)

type Message struct {
	ID        int64  `json:"id"`
	Sender    string `json:"sender"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	IsAdmin   bool   `json:"isAdmin"`
}

var (
	messages []Message
	mu       sync.Mutex

	// Telemetry
	msgCounter int
	netManager *network.NetworkManager
	govManager *governance.GovernanceManager
)

func reportMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		mu.Lock()
		rate := msgCounter
		msgCounter = 0
		mu.Unlock()

		if netManager != nil {
			netManager.UpdateDeviceMetrics("svc-chat", network.DeviceMetrics{
				RequestsPerSec: float64(rate),
				LastActivity:   time.Now().Unix(),
				Custom: map[string]interface{}{
					"messages_per_sec": rate,
					"total_messages":   len(messages),
				},
			})
			netManager.UpdateServiceMetrics("chat", map[string]interface{}{
				"messages_per_sec": rate,
				"total_messages":   len(messages),
			})
		}
	}
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

func handleMessages(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	start := 0
	if len(messages) > 50 {
		start = len(messages) - 50
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages[start:])
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	mu.Lock()
	defer mu.Unlock()

	// Governance Check: Rate Limit
	if govManager != nil {
		policy := govManager.PolicyEngine.GetPolicy()
		if msgCounter >= policy.ChatRateLimit { // Use >= to check if current count meets or exceeds limit
			govManager.ReportEvent("Spam", governance.LevelWarning,
				"Chat spam detected",
				fmt.Sprintf("Messages per second exceeds policy limit (%d)", policy.ChatRateLimit),
				"Throttling connections")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
	}

	msg.ID = time.Now().UnixNano()
	msg.Timestamp = time.Now().Format("15:04:05")
	if msg.Sender == "" {
		msg.Sender = "Anonymous"
	}
	messages = append(messages, msg)
	msgCounter++
	if len(messages) > 1000 {
		messages = messages[1:]
	}

	utils.LogInfo("Chat", fmt.Sprintf("[%s]: %s", msg.Sender, msg.Content))
	w.WriteHeader(201)
	json.NewEncoder(w).Encode(msg)
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("chat").Parse(ChatHTML)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	data := map[string]interface{}{
		"LocalIP": utils.GetLocalIP(),
		"Port":    config.ChatPort,
	}
	tmpl.Execute(w, data)
}

func Start(nm *network.NetworkManager, gm *governance.GovernanceManager) {
	netManager = nm
	govManager = gm
	messages = append(messages, Message{
		ID:        time.Now().UnixNano(),
		Sender:    "System",
		Content:   "Quantum Encryption Tunnel Established. Secure Chat Active.",
		Timestamp: time.Now().Format("15:04:05"),
		IsAdmin:   true,
	})

	go reportMetrics()

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleUI)
	mux.HandleFunc("/api/messages", enableCORS(handleMessages))
	mux.HandleFunc("/api/send", enableCORS(handleSend))

	localIP := utils.GetLocalIP()
	utils.LogInfo("Chat", fmt.Sprintf("Web Interface:     http://%s:%s", localIP, config.ChatPort))
	utils.SaveEndpoint("chat", fmt.Sprintf("http://%s:%s", localIP, config.ChatPort))

	server := &http.Server{
		Addr:    ":" + config.ChatPort,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.LogFatal("Chat", err.Error())
	}
}

const ChatHTML = `<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NEXA | Quantum Chat</title>
    <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;600;800&family=Cairo:wght@400;600;700;900&display=swap" rel="stylesheet">
    <link href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css" rel="stylesheet">
    <style>
        :root {
            --primary: #6366f1;
            --secondary: #ec4899;
            --bg: #020617;
            --card-bg: rgba(15, 23, 42, 0.8);
            --border: rgba(255, 255, 255, 0.1);
            --text: #f8fafc;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: 'Outfit', 'Cairo', sans-serif;
            background: var(--bg);
            background-image: radial-gradient(circle at 0% 0%, rgba(99, 102, 241, 0.15) 0%, transparent 50%);
            color: var(--text);
            height: 100vh;
            display: flex;
            flex-direction: column;
            overflow: hidden;
        }
        header {
            padding: 20px 40px;
            background: rgba(15, 23, 42, 0.6);
            backdrop-filter: blur(20px);
            border-bottom: 1px solid var(--border);
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .logo-box { display: flex; align-items: center; gap: 15px; }
        .logo { font-size: 1.5rem; font-weight: 800; background: linear-gradient(to right, var(--primary), var(--secondary)); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
        #chat-window {
            flex: 1;
            padding: 40px;
            overflow-y: auto;
            display: flex;
            flex-direction: column;
            gap: 16px;
        }
        .message {
            max-width: 70%;
            padding: 16px 20px;
            border-radius: 20px;
            position: relative;
            animation: fadeIn 0.3s ease;
        }
        .message.system { align-self: center; background: rgba(99, 102, 241, 0.1); border: 1px solid rgba(99, 102, 241, 0.3); font-size: 0.9rem; color: var(--primary); }
        .message.user { align-self: flex-start; background: var(--card-bg); border-radius: 4px 20px 20px 20px; border: 1px solid var(--border); }
        .message.mine { align-self: flex-end; background: linear-gradient(135deg, var(--primary), var(--secondary)); border-radius: 20px 20px 4px 20px; }
        .sender { font-size: 0.75rem; font-weight: 700; margin-bottom: 4px; opacity: 0.8; }
        .content { font-size: 1rem; line-height: 1.5; }
        .time { font-size: 0.65rem; margin-top: 6px; opacity: 0.6; text-align: left; }
        footer {
            padding: 30px 40px;
            background: rgba(15, 23, 42, 0.6);
            border-top: 1px solid var(--border);
            display: flex;
            gap: 20px;
        }
        input {
            flex: 1;
            background: rgba(0,0,0,0.3);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 16px 24px;
            color: white;
            font-family: inherit;
            outline: none;
            transition: all 0.3s;
        }
        input:focus { border-color: var(--primary); background: rgba(0,0,0,0.5); }
        button {
            padding: 0 32px;
            border-radius: 16px;
            border: none;
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            color: white;
            font-weight: 700;
            cursor: pointer;
            transition: all 0.3s;
        }
        button:hover { transform: translateY(-2px); box-shadow: 0 10px 20px -5px rgba(99, 102, 241, 0.4); }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }
    </style>
</head>
<body>
    <header>
        <div class="logo-box">
            <span style="color: white; font-size: 1.2rem; opacity: 0.7; display: flex; align-items: center;"><i class="fas fa-th-large"></i></span>
            <div class="logo">NEXA CHAT v3.1</div>
        </div>
        <div id="status" style="font-size: 0.8rem; color: #4ade80;"><i class="fas fa-circle"></i> متصل بالألياف البصرية</div>
    </header>
    <div id="chat-window"></div>
    <footer>
        <input type="text" id="msg-input" placeholder="اكتب رسالتك المشفرة هنا..." autocomplete="off">
        <button onclick="sendMessage()">إرسال <i class="fas fa-paper-plane"></i></button>
    </footer>
    <script>
        const chatWindow = document.getElementById('chat-window');
        const msgInput = document.getElementById('msg-input');
        let lastId = 0;
        async function fetchMessages() {
            try {
                const resp = await fetch('/api/messages');
                const msgs = await resp.json();
                msgs.forEach(m => {
                    if (m.id > lastId) {
                        appendMessage(m);
                        lastId = m.id;
                    }
                });
            } catch (e) {}
        }
        function appendMessage(m) {
            const div = document.createElement('div');
            div.className = 'message ' + (m.sender === 'System' ? 'system' : 'user');
            div.innerHTML = (m.sender !== 'System' ? '<div class="sender">' + m.sender + '</div>' : '') + '<div class="content">' + m.content + '</div><div class="time">' + m.timestamp + '</div>';
            chatWindow.appendChild(div);
            chatWindow.scrollTop = chatWindow.scrollHeight;
        }
        async function sendMessage() {
            const text = msgInput.value.trim();
            if (!text) return;
            msgInput.value = '';
            try {
                await fetch('/api/send', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify({sender: 'User', content: text})
                });
                fetchMessages();
            } catch (e) {}
        }
        msgInput.addEventListener('keydown', e => { if (e.key === 'Enter') sendMessage(); });
        setInterval(fetchMessages, 2000);
        fetchMessages();
    </script>
</body>
</html>`
