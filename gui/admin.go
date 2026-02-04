package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
)

type LogEntry struct {
	Timestamp time.Time
	Username  string
	Command   string
	Result    string
}

var logs []LogEntry
var logMutex sync.Mutex

var loginPage = `
<!DOCTYPE html>
<html lang="ar">
<head>
<meta charset="UTF-8">
<title>تسجيل الدخول - Nexa</title>
<style>
body { font-family: Tahoma, Arial, sans-serif; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); margin: 0; height: 100vh; display: flex; align-items: center; justify-content: center; }
.container { max-width: 400px; background: #fff; border-radius: 12px; box-shadow: 0 10px 25px #0003; padding: 40px; }
h1 { color: #667eea; text-align: center; }
input { width: 100%; margin: 12px 0 16px 0; padding: 12px; border-radius: 6px; border: 1px solid #ddd; box-sizing: border-box; }
button { width: 100%; background: #667eea; color: #fff; border: none; padding: 12px; border-radius: 6px; font-size: 1em; cursor: pointer; margin-top: 10px; }
button:hover { background: #764ba2; }
.err { color: #e74c3c; margin-bottom: 15px; text-align: center; background: #fadbd8; padding: 10px; border-radius: 6px; }
.info { text-align: center; color: #7f8c8d; font-size: 0.9em; margin-top: 20px; }
</style>
</head>
<body>
<div class="container">
<h1>Nexa - تسجيل الدخول</h1>
{{if .Error}}<div class="err">{{.Error}}</div>{{end}}
<form method="POST">
<input name="user" placeholder="اسم المستخدم" required autocomplete="off">
<input name="pass" type="password" placeholder="كلمة المرور" required>
<button type="submit">دخول</button>
</form>
<div class="info">
<p>بيانات تجريبية:</p>
<p>admin / admin123<br>user1 / admin123<br>guest / admin123</p>
</div>
</div>
</body>
</html>
`

var dashboardPage = `
<!DOCTYPE html>
<html lang="ar">
<head>
<meta charset="UTF-8">
<title>لوحة التحكم - Nexa</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: Tahoma, Arial, sans-serif; background: #f5f7fa; }
.header { background: #667eea; color: #fff; padding: 20px; text-align: center; }
.nav { background: #555; overflow: auto; }
.nav a, .nav button { float: left; color: #fff; text-align: center; padding: 14px 20px; text-decoration: none; cursor: pointer; border: none; font-size: 0.9em; }
.nav a:hover, .nav button:hover { background: #764ba2; }
.nav button { float: right; background: #e74c3c; }
.content { max-width: 900px; margin: 20px auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #0001; padding: 24px; }
.tab { display: none; }
.tab.active { display: block; }
h2 { color: #667eea; margin-bottom: 20px; }
input, textarea, select { width: 100%; margin: 8px 0 16px 0; padding: 10px; border-radius: 4px; border: 1px solid #ddd; }
button.btn { background: #667eea; color: #fff; border: none; padding: 10px 20px; border-radius: 4px; cursor: pointer; }
button.btn:hover { background: #764ba2; }
button.del { background: #e74c3c; }
table { width: 100%; border-collapse: collapse; margin-top: 15px; }
table th, table td { border: 1px solid #ddd; padding: 12px; text-align: right; }
table th { background: #f0f0f0; }
pre { background: #222; color: #0f0; padding: 12px; border-radius: 4px; overflow-x: auto; height: 300px; overflow-y: auto; }
</style>
</head>
<body>
<div class="header">
<h1>Nexa - لوحة التحكم</h1>
<p>مرحباً {{.Username}} ({{.Role}})</p>
</div>
<div class="nav">
<a href="#" onclick="switchTab('dashboard')">لوحة التحكم</a>
{{if eq .Role "admin"}}
<a href="#" onclick="switchTab('users')">إدارة المستخدمين</a>
{{end}}
<a href="#" onclick="switchTab('commands')">إرسال أوامر</a>
<a href="#" onclick="switchTab('logs')">السجلات</a>
<button onclick="logout()">تسجيل خروج</button>
</div>

<div class="content">
<!-- Dashboard Tab -->
<div id="dashboard" class="tab active">
<h2>ملخص النظام</h2>
<p>عدد المستخدمين: {{.UserCount}}</p>
<p>عدد السجلات: {{.LogCount}}</p>
<p>آخر نشاط: {{.LastActivity}}</p>
</div>

<!-- Users Management Tab (Admin Only) -->
{{if eq .Role "admin"}}
<div id="users" class="tab">
<h2>إدارة المستخدمين</h2>
<form method="POST" action="/admin/add-user">
<h3>إضافة مستخدم جديد</h3>
<input name="user" placeholder="اسم المستخدم" required>
<input name="pass" type="password" placeholder="كلمة المرور" required>
<select name="role">
<option value="user">مستخدم عادي</option>
<option value="admin">إدارة</option>
<option value="guest">زائر</option>
</select>
<button class="btn" type="submit">إضافة</button>
</form>

<h3>المستخدمون الحاليون</h3>
<table>
<tr><th>الاسم</th><th>الدور</th><th>الإجراء</th></tr>
{{range .Users}}<tr><td>{{.Name}}</td><td>{{.Role}}</td><td><form method="POST" action="/admin/del-user" style="display:inline;"><input type="hidden" name="user" value="{{.Name}}"><button class="btn del" type="submit">حذف</button></form></td></tr>{{end}}
</table>
</div>
{{end}}

<!-- Commands Tab -->
<div id="commands" class="tab">
<h2>إرسال أوامر Nexa</h2>
<form method="POST" action="/command">
<textarea name="cmd" rows="4" placeholder="مثال: PING أو FETCH homepage أو PUBLISH homepage مرحباً"></textarea>
<label><input type="checkbox" name="dns"> استخدم DNS (.nexa)</label>
<button class="btn" type="submit">إرسال</button>
</form>
{{if .CommandResult}}<h3>النتيجة:</h3><pre>{{.CommandResult}}</pre>{{end}}
</div>

<!-- Logs Tab -->
<div id="logs" class="tab">
<h2>سجل العمليات</h2>
<table>
<tr><th>الوقت</th><th>المستخدم</th><th>الأمر</th><th>النتيجة</th></tr>
{{range .Logs}}<tr><td>{{.Timestamp.Format "2006-01-02 15:04:05"}}</td><td>{{.Username}}</td><td>{{.Command}}</td><td>{{.Result}}</td></tr>{{end}}
</table>
</div>
</div>

<script>
function switchTab(tabName) {
  var tabs = document.getElementsByClassName("tab");
  for (var i = 0; i < tabs.length; i++) {
    tabs[i].classList.remove("active");
  }
  document.getElementById(tabName).classList.add("active");
}
function logout() {
  window.location.href = '/logout';
}
</script>
</body>
</html>
`

type pageData struct {
	Username      string
	Role          string
	UserCount     int
	LogCount      int
	LastActivity  string
	Users         []map[string]string
	Logs          []LogEntry
	CommandResult string
	Error         string
}

var users map[string]struct {
	Password string `json:"password"`
	Role     string `json:"role"`
}

var sessions = map[string]string{} // sessionID -> username

func main() {
	loadUsers()
	http.HandleFunc("/", authHandler(dashboardHandler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/command", authHandler(commandHandler))
	http.HandleFunc("/admin/add-user", adminHandler(addUserHandler))
	http.HandleFunc("/admin/del-user", adminHandler(delUserHandler))
	fmt.Println("Nexa Admin Panel running at http://0.0.0.0:8080 ...")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	username := getUsername(r)
	role := users[username].Role
	data := pageData{
		Username:     username,
		Role:         role,
		UserCount:    len(users),
		LogCount:     len(logs),
		LastActivity: getLastActivity(),
		Users:        getUsersList(),
		Logs:         getLast10Logs(),
	}
	tmpl := template.Must(template.New("dashboard").Parse(dashboardPage))
	tmpl.Execute(w, data)
}

func commandHandler(w http.ResponseWriter, r *http.Request) {
	username := getUsername(r)
	role := users[username].Role
	data := pageData{
		Username: username,
		Role:     role,
	}
	if r.Method == http.MethodPost {
		cmd := r.FormValue("cmd")
		useDNS := r.FormValue("dns") == "on"
		var serverAddr string
		if useDNS {
			parts := strings.Fields(cmd)
			if len(parts) < 2 {
				data.CommandResult = "الأمر يحتاج اسم .nexa!"
			} else {
				addr, err := resolveDNS(parts[1])
				if err != nil {
					data.CommandResult = "فشل حل DNS: " + err.Error()
				} else {
					serverAddr = addr
					resp, err := sendCommand(serverAddr, cmd)
					if err != nil {
						data.CommandResult = "خطأ: " + err.Error()
					} else {
						data.CommandResult = resp
					}
				}
			}
		} else {
			resp, err := sendCommand("localhost:1413", cmd)
			if err != nil {
				data.CommandResult = "خطأ: " + err.Error()
			} else {
				data.CommandResult = resp
			}
		}
		addLog(username, cmd, data.CommandResult)
	}
	tmpl := template.Must(template.New("dashboard").Parse(dashboardPage))
	tmpl.Execute(w, data)
}

func addUserHandler(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	role := r.FormValue("role")
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	users[user] = struct {
		Password string `json:"password"`
		Role     string `json:"role"`
	}{string(hashed), role}
	saveUsers()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func delUserHandler(w http.ResponseWriter, r *http.Request) {
	user := r.FormValue("user")
	if user != "admin" {
		delete(users, user)
		saveUsers()
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	data := pageData{}
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
		data.Error = "بيانات الدخول غير صحيحة"
	}
	tmpl := template.Must(template.New("login").Parse(loginPage))
	tmpl.Execute(w, data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("sid")
	if err == nil {
		delete(sessions, c.Value)
		http.SetCookie(w, &http.Cookie{Name: "sid", Value: "", Path: "/", MaxAge: -1})
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func getUsername(r *http.Request) string {
	c, _ := r.Cookie("sid")
	return sessions[c.Value]
}

func loadUsers() {
	f, err := os.Open("../users.json")
	if err != nil {
		log.Fatalf("users.json not found: %v", err)
	}
	defer f.Close()
	json.NewDecoder(f).Decode(&users)
}

func saveUsers() {
	f, _ := os.Create("../users.json")
	defer f.Close()
	json.NewEncoder(f).Encode(users)
}

func newSessionID() string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	sid := make([]rune, 32)
	for i := range sid {
		sid[i] = letters[rand.Intn(len(letters))]
	}
	return string(sid)
}

func sendCommand(addr, cmd string) (string, error) {
	conn, err := tls.Dial("tcp", addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	fmt.Fprintf(conn, "%s\n", cmd)
	var sb strings.Builder
	reader := make([]byte, 4096)
	for {
		n, err := conn.Read(reader)
		sb.Write(reader[:n])
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
	conn, err := tls.Dial("tcp", "localhost:1112", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	fmt.Fprintf(conn, "RESOLVE %s\n", name)
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	resp := string(buf[:n])
	parts := strings.SplitN(resp, " ", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("استجابة DNS غير صالحة")
	}
	if parts[0] != "200" {
		return "", fmt.Errorf(resp)
	}
	body := parts[2]
	addr := strings.Split(body, "|")[0]
	return addr, nil
}

func addLog(username, cmd, result string) {
	logMutex.Lock()
	defer logMutex.Unlock()
	logs = append(logs, LogEntry{time.Now(), username, cmd, result[:50]})
}

func getLast10Logs() []LogEntry {
	logMutex.Lock()
	defer logMutex.Unlock()
	if len(logs) <= 10 {
		return logs
	}
	return logs[len(logs)-10:]
}

func getUsersList() []map[string]string {
	var list []map[string]string
	for name, u := range users {
		list = append(list, map[string]string{"Name": name, "Role": u.Role})
	}
	return list
}

func getLastActivity() string {
	logMutex.Lock()
	defer logMutex.Unlock()
	if len(logs) == 0 {
		return "لا توجد أنشطة"
	}
	return logs[len(logs)-1].Timestamp.Format("2006-01-02 15:04:05")
}
