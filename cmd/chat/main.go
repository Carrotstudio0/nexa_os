package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

const ChatPort = "8082"

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
)

// CORS Middleware
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
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

func handleMessages(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	// Return last 50 messages
	start := 0
	if len(messages) > 50 {
		start = len(messages) - 50
	}

	json.NewEncoder(w).Encode(messages[start:])
}

func handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	msg.ID = time.Now().UnixNano()
	msg.Timestamp = time.Now().Format("15:04") // HH:MM format

	// Basic validation
	if msg.Sender == "" {
		msg.Sender = "Anonymous"
	}
	if msg.Content == "" {
		mu.Unlock()
		return
	}

	messages = append(messages, msg)
	// Keep memory clean - limit to 1000 messages
	if len(messages) > 1000 {
		messages = messages[1:]
	}
	mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

func handleClear(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	messages = []Message{{
		ID:        time.Now().UnixNano(),
		Sender:    "System",
		Content:   "Chat history cleared",
		Timestamp: time.Now().Format("15:04"),
		IsAdmin:   true,
	}}
	mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func main() {
	// Initial welcome message
	messages = append(messages, Message{
		ID:        time.Now().UnixNano(),
		Sender:    "System",
		Content:   "Welcome to Nexa Chat! 💬",
		Timestamp: time.Now().Format("15:04"),
		IsAdmin:   true,
	})

	http.HandleFunc("/messages", enableCORS(handleMessages))
	http.HandleFunc("/send", enableCORS(handleSend))
	http.HandleFunc("/clear", enableCORS(handleClear))

	localIP := getLocalIP()
	fmt.Println(`
      ____ _           _      ____                  _          
     / ___| |__   __ _| |_   / ___|  ___ _ ____   _(_) ___ ___ 
    | |   | '_ \ / _' | __|  \___ \ / _ \ '__\ \ / / |/ __/ _ \
    | |___| | | | (_| | |_    ___) |  __/ |   \ V /| | (_|  __/
     \____|_| |_|\__,_|\__|  |____/ \___|_|    \_/ |_|\___\___|
                                                  v3.0 Ultimate`)
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Printf("   [INFO]  Initializing Chat System...\n")
	fmt.Printf("   [INFO]  API Endpoint:      http://%s:%s\n", localIP, ChatPort)
	fmt.Printf("   [INFO]  History Limit:     %d Messages\n", 1000)
	fmt.Printf("   [INFO]  CORS Policy:       %s\n", "ENABLED (Global)")
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Println("   ✅  CHAT SERVER READY")

	if err := http.ListenAndServe(":"+ChatPort, nil); err != nil {
		log.Fatal(err)
	}
}
