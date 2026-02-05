package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MultiX0/nexa/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

// --- Configuration ---
const (
	AdminPort = "8080"
	ServerURL = "localhost:1413"
	DNSURL    = "localhost:1112"
)

// --- Data Structures ---
type LogEntry struct {
	Timestamp time.Time
	Username  string
	Command   string
	Result    string
	Status    string // SUCCESS, ERROR
}

var (
	logs     []LogEntry
	logMutex sync.Mutex
	users    map[string]User
	sessions = map[string]string{} // sessionID -> username
)

type User struct {
	Password string `json:"password"`
	Role     string `json:"role"`
}

var usersFilePath string

// --- Main ---
func main() {
	loadUsers()

	// Static Assets (if any, using embedded CSS in template for now)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Routes
	http.HandleFunc("/", authHandler(dashboardHandler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/api/command", authHandler(apiCommandHandler)) // AJAX handler
	http.HandleFunc("/admin/users", adminHandler(usersHandler))

	// Professional Startup Banner
	fmt.Println(`
       _       _           _         ____                  _ 
      / \   __| |_ __ ___ (_)_ __   |  _ \ __ _ _ __   ___| |
     / _ \ / _' | '_ ' _ \| | '_ \  | |_) / _' | '_ \ / _ \ |
    / ___ \ (_| | | | | | | | | | | |  __/ (_| | | | |  __/ |
   /_/   \_\__,_|_| |_| |_|_|_| |_| |_|   \__,_|_| |_|\___|_|
                                               v3.0 Ultimate`)
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Printf("   [INFO]  Initializing Admin Controller...\n")
	fmt.Printf("   [INFO]  Control Panel:     http://localhost:%s\n", AdminPort)
	fmt.Printf("   [INFO]  User Database:     %d users loaded\n", len(users))
	fmt.Printf("   [INFO]  Security Level:    %s\n", "HIGH (Bcrypt Enabled)")
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Println("   ✅  ADMIN PANEL READY")

	log.Fatal(http.ListenAndServe("0.0.0.0:"+AdminPort, nil))
}

// --- Handlers ---

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	username := getUsername(r)
	role := users[username].Role

	data := map[string]interface{}{
		"Username":     username,
		"Role":         role,
		"UserCount":    len(users),
		"LogCount":     len(logs),
		"SystemStatus": "OPERATIONAL",
	}

	tmpl := template.Must(template.New("layout").Parse(LayoutHTML))
	tmpl.Execute(w, data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		user := r.FormValue("user")
		pass := r.FormValue("pass")
		if u, ok := users[user]; ok {
			if bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass)) == nil {
				sid := newSessionID()
				sessions[sid] = user
				http.SetCookie(w, &http.Cookie{Name: "sid", Value: sid, Path: "/", HttpOnly: true})
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
		}
		// Show error
		tmpl := template.Must(template.New("login").Parse(LoginHTML))
		tmpl.Execute(w, map[string]string{"Error": "بيانات الدخول غير صحيحة"})
		return
	}
	tmpl := template.Must(template.New("login").Parse(LoginHTML))
	tmpl.Execute(w, nil)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("sid")
	if err == nil {
		delete(sessions, c.Value)
		http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", Path: "/", MaxAge: -1})
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func apiCommandHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", 405)
		return
	}

	cmd := r.FormValue("cmd")
	useDNS := r.FormValue("dns") == "true"
	username := getUsername(r)

	var result string
	var err error
	var status = "SUCCESS"

	if useDNS {
		// Logic to resolve and send
		parts := strings.Fields(cmd)
		if len(parts) < 2 || !strings.HasSuffix(parts[1], ".nexa") {
			result = "خطأ: يجب تحديد نطاق .nexa عند استخدام DNS"
			status = "ERROR"
		} else {
			addr, errDNS := resolveDNS(parts[1])
			if errDNS != nil {
				result = "DNS Error: " + errDNS.Error()
				status = "ERROR"
			} else {
				result, err = sendCommand(addr, cmd)
			}
		}
	} else {
		result, err = sendCommand(ServerURL, cmd)
	}

	if err != nil {
		result = "Connection Error: " + err.Error()
		status = "ERROR"
	}

	addLog(username, cmd, result, status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"result": result,
		"status": status,
		"time":   time.Now().Format("15:04:05"),
	})
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		action := r.FormValue("action")
		if action == "add" {
			u := r.FormValue("username")
			p := r.FormValue("password")
			role := r.FormValue("role")
			hashed, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
			users[u] = User{Password: string(hashed), Role: role}
			saveUsers()
		} else if action == "delete" {
			u := r.FormValue("username")
			if u != "admin" {
				delete(users, u)
				saveUsers()
			}
		}
	}
	// Return updated list JSON for dynamic UI? Or just redirect.
	// For this simple version, we'll redirect.
	http.Redirect(w, r, "/?tab=users", http.StatusSeeOther)
}

// --- Helpers ---

func getUsername(r *http.Request) string {
	c, _ := r.Cookie("sid")
	return sessions[c.Value]
}

func authHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("sid")
		if err != nil || sessions[c.Value] == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func adminHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("sid")
		if err != nil || sessions[c.Value] == "" || users[sessions[c.Value]].Role != "admin" {
			http.Error(w, "Unauthorized", 403)
			return
		}
		next(w, r)
	}
}

func loadUsers() {
	usersFilePath = utils.FindFile("users.json")
	f, err := os.Open(usersFilePath)
	if err != nil {
		users = map[string]User{
			"admin": {Password: "$2a$10$N9qo8uLOickgx2ZMRZoHK.ZG8rHv5yPXrOqQ5qM0jPPEYLRKZMiMO", Role: "admin"}, // admin123
			"user1": {Password: "$2a$10$N9qo8uLOickgx2ZMRZoHK.ZG8rHv5yPXrOqQ5qM0jPPEYLRKZMiMO", Role: "user"},
		}
		usersFilePath = "users.json"
		saveUsers()
		return
	}
	defer f.Close()
	json.NewDecoder(f).Decode(&users)
}

func saveUsers() {
	f, _ := os.Create(usersFilePath)
	defer f.Close()
	json.NewEncoder(f).Encode(users)
}

func newSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func addLog(user, cmd, res, status string) {
	logMutex.Lock()
	defer logMutex.Unlock()
	logs = append(logs, LogEntry{time.Now(), user, cmd, res, status})
	if len(logs) > 50 {
		logs = logs[1:]
	}
}

// -- Networking --

func sendCommand(addr, cmd string) (string, error) {
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	fmt.Fprintf(conn, "%s\n", cmd)
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			return sb.String(), err
		}
		if strings.Contains(sb.String(), "---END---") {
			break
		}
	}
	return sb.String(), nil
}

func resolveDNS(name string) (string, error) {
	conn, err := tls.Dial("tcp", DNSURL, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	fmt.Fprintf(conn, "RESOLVE %s\n", name)
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	resp := string(buf[:n])
	// Parse: 200 RESOLVED ip:port|...
	parts := strings.SplitN(resp, " ", 3)
	if len(parts) < 3 || parts[0] != "200" {
		return "", fmt.Errorf("Resolve failed")
	}
	return strings.Split(parts[2], "|")[0], nil
}
