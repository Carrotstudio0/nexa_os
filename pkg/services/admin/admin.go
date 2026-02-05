package admin

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	AdminPort = config.AdminPort
	ServerURL = "localhost:" + config.ServerPort
	DNSURL    = "localhost:" + config.DNSPort
)

type LogEntry struct {
	Timestamp time.Time
	Username  string
	Command   string
	Result    string
	Status    string
}

var (
	logs       []LogEntry
	logMutex   sync.Mutex
	users      map[string]User
	sessions   = map[string]string{}
	netManager *network.NetworkManager
	govManager *governance.GovernanceManager
)

type User struct {
	Password string `json:"password"`
	Role     string `json:"role"`
}

var usersFilePath string

func Start(nm *network.NetworkManager, gm *governance.GovernanceManager) {
	netManager = nm
	govManager = gm
	loadUsers()

	// Metrics reporter
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			if netManager != nil {
				netManager.UpdateServiceMetrics("admin", map[string]interface{}{
					"log_count":      len(logs),
					"active_session": len(sessions),
					"user_count":     len(users),
				})
			}
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", authHandler(dashboardHandler))
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/api/command", authHandler(apiCommandHandler))
	mux.HandleFunc("/admin/users", adminHandler(usersHandler))

	utils.LogInfo("Admin", "Unified Service Starting...")
	utils.SaveEndpoint("admin", fmt.Sprintf("http://%s:%s", utils.GetLocalIP(), AdminPort))

	server := &http.Server{
		Addr:    "0.0.0.0:" + AdminPort,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		utils.LogFatal("Admin", err.Error())
	}
}

// ... (Rest of the handlers updated to use 'mux' or just left as is if they don't depend on global mux)
// Actually I need to move ALL the code here.

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
		switch action {
		case "add":
			u := r.FormValue("username")
			p := r.FormValue("password")
			role := r.FormValue("role")
			hashed, _ := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
			users[u] = User{Password: string(hashed), Role: role}
			saveUsers()
		case "delete":
			u := r.FormValue("username")
			if u != "admin" {
				delete(users, u)
				saveUsers()
			}
		}
	}
	http.Redirect(w, r, "/?tab=users", http.StatusSeeOther)
}

func getUsername(r *http.Request) string {
	c, _ := r.Cookie("sid")
	if c == nil {
		return ""
	}
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
	usersFilePath = "users.json"
	f, err := os.Open(usersFilePath)
	if err != nil {
		users = map[string]User{
			"admin": {Password: "$2a$10$N9qo8uLOickgx2ZMRZoHK.ZG8rHv5yPXrOqQ5qM0jPPEYLRKZMiMO", Role: "admin"},
			"user1": {Password: "$2a$10$N9qo8uLOickgx2ZMRZoHK.ZG8rHv5yPXrOqQ5qM0jPPEYLRKZMiMO", Role: "user"},
		}
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
	parts := strings.SplitN(resp, " ", 3)
	if len(parts) < 3 || parts[0] != "200" {
		return "", fmt.Errorf("Resolve failed")
	}
	return strings.Split(parts[2], "|")[0], nil
}
