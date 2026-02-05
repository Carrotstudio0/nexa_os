package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/MultiX0/nexa/pkg/services/admin"
	"github.com/MultiX0/nexa/pkg/services/chat"
	"github.com/MultiX0/nexa/pkg/services/dashboard"
	"github.com/MultiX0/nexa/pkg/services/dns"
	"github.com/MultiX0/nexa/pkg/services/gateway"
	"github.com/MultiX0/nexa/pkg/services/server"
	"github.com/MultiX0/nexa/pkg/services/storage"
	"github.com/MultiX0/nexa/pkg/utils"
)

func main() {
	utils.PrintBanner("NEXA ULTIMATE CORE", "v3.1")
	fmt.Println("\n[SYSTEM] Initializing Unified Service Matrix...")

	var wg sync.WaitGroup

	// Start all services in goroutines
	services := []struct {
		Name  string
		Start func()
	}{
		{"Core Server", server.Start},
		{"DNS Authority", dns.Start},
		{"File Storage", storage.Start},
		{"Quantum Chat", chat.Start},
		{"Admin Panel", admin.Start},
		{"Network Gateway", gateway.Start},
		{"System Dashboard", dashboard.Start},
	}

	for _, s := range services {
		wg.Add(1)
		go func(name string, startFunc func()) {
			defer wg.Done()
			utils.LogInfo("Matrix", "Launching "+name+"...")
			startFunc()
		}(s.Name, s.Start)
	}

	utils.LogSuccess("System", "ALL SERVICES OPERATIONAL")

	localIP := utils.GetLocalIP()
	fmt.Printf("\n  +--------------------------------------------------+\n")
	fmt.Printf("  |  🌐 UNIFIED HUB: http://%s:7000        |\n", localIP)
	fmt.Printf("  |  🚪 GATEWAY:     http://%s:8000        |\n", localIP)
	fmt.Printf("  +--------------------------------------------------+\n\n")

	// Wait for interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	utils.LogInfo("System", "Nexa is running. Press Ctrl+C to shutdown.")
	<-stop

	fmt.Println("\n[SYSTEM] Graceful shutdown initiated...")
	// In a real app we'd trigger service shutdowns, but for now we exit.
	os.Exit(0)
}
