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
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"time"
)

var loginPage = `
<!DOCTYPE html>
<html lang="ar">
<head>
<meta charset="UTF-8">
<title>تسجيل الدخول - Nexa</title>
<style>
body { font-family: Tahoma, Arial, sans-serif; background: #f7f7f7; }
.container { max-width: 400px; margin: 60px auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #0001; padding: 32px; }
h1 { color: #2a4d7a; }
input { width: 100%; margin: 8px 0 16px 0; padding: 8px; border-radius: 4px; border: 1px solid #ccc; }
button { background: #2a4d7a; color: #fff; border: none; padding: 10px 24px; border-radius: 4px; font-size: 1em; cursor: pointer; }
button:hover { background: #183153; }
.err { color: #b00; margin-bottom: 10px; }
</style>
</head>
<body>
<div class="container">
<h1>تسجيل الدخول</h1>
{{if .Error}}<div class="err">{{.Error}}</div>{{end}}
<form method="POST">
<input name="user" placeholder="اسم المستخدم" required autocomplete="off">
<input name="pass" type="password" placeholder="كلمة المرور" required>
<button type="submit">دخول</button>
</form>
</div>
</body>
</html>
`

var page = `
<!DOCTYPE html>
<html lang="ar">
<head>
<meta charset="UTF-8">
<title>Nexa Web Client</title>
<style>
body { font-family: Tahoma, Arial, sans-serif; background: #f7f7f7; margin: 0; padding: 0; }
.container { max-width: 600px; margin: 40px auto; background: #fff; border-radius: 8px; box-shadow: 0 2px 8px #0001; padding: 32px; }
h1 { color: #2a4d7a; }
label { font-weight: bold; }
input, textarea, select { width: 100%; margin: 8px 0 16px 0; padding: 8px; border-radius: 4px; border: 1px solid #ccc; }
button { background: #2a4d7a; color: #fff; border: none; padding: 10px 24px; border-radius: 4px; font-size: 1em; cursor: pointer; }
button:hover { background: #183153; }
pre { background: #222; color: #eee; padding: 12px; border-radius: 4px; overflow-x: auto; }
</style>
</head>
<body>
<div class="container">
<h1>Nexa Web Client</h1>
<form method="POST">
<label>الأمر (مثال: PING أو FETCH homepage أو PUBLISH homepage مرحباً):</label>
<input name="cmd" required autocomplete="off" placeholder="مثال: PING أو FETCH homepage ...">
<label><input type="checkbox" name="dns"> استخدم DNS (.nexa)</label>
<button type="submit">إرسال</button>
</form>
{{if .Result}}
<h3>النتيجة:</h3>
<pre>{{.Result}}</pre>
{{end}}
</div>
</body>
</html>
`

type pageData struct {
	Result string
	Error  string
}

var users map[string]struct {
	Password string `json:"password"`
	Role     string `json:"role"`
}

var sessions = map[string]string{} // sessionID -> username

func main() {
	loadUsers()
	http.HandleFunc("/", authHandler(handler))
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	fmt.Println("Nexa Web Client running at http://localhost:8080 ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	data := pageData{}
	if r.Method == http.MethodPost {
		r.ParseForm()
		cmd := r.FormValue("cmd")
		useDNS := r.FormValue("dns") == "on"
		var serverAddr string
		if useDNS {
			parts := strings.Fields(cmd)
			if len(parts) < 2 {
				data.Result = "الأمر يحتاج اسم .nexa!"
				render(w, data)
				return
			}
			name := parts[1]
			addr, err := resolveDNS(name)
			if err != nil {
				data.Result = "فشل حل DNS: " + err.Error()
				render(w, data)
				return
			}
			serverAddr = addr
		} else {
			serverAddr = "localhost:1413"
		}
		resp, err := sendCommand(serverAddr, cmd)
		if err != nil {
			data.Result = "خطأ: " + err.Error()
		} else {
			data.Result = resp
		}
	}
	render(w, data)
}

func render(w http.ResponseWriter, data pageData) {
	tmpl := template.Must(template.New("page").Parse(page))
	tmpl.Execute(w, data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	data := pageData{}
	if r.Method == http.MethodPost {
		r.ParseForm()
		user := r.FormValue("user")
		pass := r.FormValue("pass")
		if u, ok := users[user]; ok {
			err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(pass))
			if err == nil {
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
		if r.URL.Path == "/login" {
			next := loginHandler
			next(w, r)
			return
		}
		c, err := r.Cookie("sid")
		if err != nil || sessions[c.Value] == "" {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func loadUsers() {
	f, err := os.Open("../users.json")
	if err != nil {
		log.Fatalf("users.json not found: %v", err)
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	err = dec.Decode(&users)
	if err != nil {
		log.Fatalf("users.json decode error: %v", err)
	}
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
	_, err = fmt.Fprintf(conn, "%s\n", cmd)
	if err != nil {
		return "", err
	}
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
	n, err := conn.Read(buf)
	if err != nil {
		return "", err
	}
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
