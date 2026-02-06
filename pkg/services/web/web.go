package web

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/utils"
)

//go:embed ui/index.html
var uiFiles embed.FS

const (
	WebPort     = config.WebPort
	StorageRoot = "./storage"
)

// Start initializes and starts the web service
func Start(nm *network.NetworkManager, gm *governance.GovernanceManager) {
	// Ensure storage directory structure
	initStorageDirectories()

	// Register routes
	http.HandleFunc("/", enableCORS(serveWebUI))
	http.HandleFunc("/api/status", enableCORS(statusHandler))

	localIP := utils.GetLocalIP()
	utils.LogInfo("Web", "v4.0.0-PRO module initialization...")
	utils.LogInfo("Web", fmt.Sprintf("Address:           http://%s:%s", localIP, WebPort))
	utils.SaveEndpoint("web", fmt.Sprintf("http://%s:%s", localIP, WebPort))

	// Start server with heartbeat
	go reportMetrics(nm)

	if err := http.ListenAndServe("0.0.0.0:"+WebPort, nil); err != nil {
		utils.LogWarning("Web", fmt.Sprintf("Server error: %v", err))
	}
}

// initStorageDirectories creates necessary storage folders
func initStorageDirectories() {
	dirs := []string{"incoming", "shared", "vault", "temp", "downloads"}
	for _, dir := range dirs {
		path := filepath.Join(StorageRoot, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.MkdirAll(path, 0755)
		}
	}
}

// serveWebUI serves the main web interface
func serveWebUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	indexHTML, err := uiFiles.ReadFile("ui/index.html")
	if err != nil {
		http.Error(w, "UI Resource Error", 500)
		return
	}
	w.Write(indexHTML)
}

// statusHandler returns the web service status
func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{
		"status": "online",
		"service": "web",
		"version": "v4.0.0-PRO",
		"port": "` + WebPort + `"
	}`))
}

// enableCORS adds CORS headers to responses
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// reportMetrics sends metrics to network manager
func reportMetrics(nm *network.NetworkManager) {
	if nm == nil {
		return
	}

	// Update service status periodically
	for {
		nm.UpdateServiceMetrics("web", map[string]interface{}{
			"status": "online",
			"port":   WebPort,
		})
		<-time.After(5 * time.Second)
	}
}
