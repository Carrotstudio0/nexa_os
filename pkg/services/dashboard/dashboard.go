package dashboard

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/utils"
)

//go:embed ui/dashboard.html
var uiFiles embed.FS

var (
	netManager *network.NetworkManager
	govManager *governance.GovernanceManager
)

func handleNetworkMap(w http.ResponseWriter, r *http.Request) {
	if netManager == nil {
		http.Error(w, "Network Manager not initialized", 503)
		return
	}

	topology := netManager.GetTopology()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(topology)
}

func handleGovernanceTimeline(w http.ResponseWriter, r *http.Request) {
	if govManager == nil {
		http.Error(w, "Governance Manager not initialized", 503)
		return
	}
	timeline := govManager.GetTimeline()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(timeline)
}

func handleGovernancePolicy(w http.ResponseWriter, r *http.Request) {
	if govManager == nil {
		http.Error(w, "Governance Manager not initialized", 503)
		return
	}

	if r.Method == http.MethodPost {
		var p governance.Policy
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			http.Error(w, "Invalid Policy Data", 400)
			return
		}
		govManager.PolicyEngine.UpdatePolicy(p)
		w.WriteHeader(200)
		return
	}

	policy := govManager.PolicyEngine.GetPolicy()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	utils.LogInfo("Dashboard", "Connection received from: "+r.RemoteAddr)
	localIP := utils.GetLocalIP()

	data := map[string]interface{}{
		"LocalIP":  localIP,
		"Port":     config.DashboardPort,
		"Services": config.Services,
	}

	tmpl, err := template.ParseFS(uiFiles, "ui/dashboard.html")
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

func Start(nm *network.NetworkManager, gm *governance.GovernanceManager) {
	netManager = nm
	govManager = gm
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleDashboard)
	mux.HandleFunc("/storage/", handleProxyFiles)
	mux.HandleFunc("/admin/", handleProxyAdmin)
	mux.HandleFunc("/chat/", handleProxyChat)
	mux.HandleFunc("/api/network/map", handleNetworkMap)
	mux.HandleFunc("/api/governance/timeline", handleGovernanceTimeline)
	mux.HandleFunc("/api/governance/policy", handleGovernancePolicy)

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
