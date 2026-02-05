package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/network"
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

	// --- Network Intelligence Layer Initialization ---
	utils.LogInfo("Network", "Initializing Knowledge Graph...")
	localIP := utils.GetLocalIP()

	netConfig := network.ConnectionConfig{
		Timeout:           5 * time.Second,
		MaxRetries:        3,
		HeartbeatInterval: 10 * time.Second,
		ConnectionType:    network.ConnectionWiFi,
	}
	nm := network.NewNetworkManager(netConfig)

	// 1. Register Primary Base (The Host)
	base, err := nm.RegisterPrimaryBase("base-core", "Nexa Core", "00:00:00:00:00:00", localIP, 7000)
	if err == nil {
		base.IsOnline = true
		base.UpdateOnlineStatus(true)
	}

	// 2. Register Service Nodes (Logical Representation)
	servicesList := []struct {
		id, name string
		port     int
	}{
		{"svc-gateway", "Gateway Node", 8000},
		{"svc-chat", "Quantum Chat", 8082},
		{"svc-storage", "Digital Storage", 8081},
		{"svc-dns", "DNS Authority", 53},
		{"svc-admin", "Admin Panel", 8080},
	}

	for _, s := range servicesList {
		node, err := nm.RegisterDevice(s.id, s.name, "", localIP, s.port, network.RoleGateway) // Using Gateway role for services for visibility
		if err == nil {
			node.IsOnline = true
			node.UpdateOnlineStatus(true)
			nm.CreateConnection(base.ID, node.ID, network.ConnectionMesh)
		}
	}

	// 3. Initialize Governance Layer
	utils.LogInfo("System", "Establishing System Constitution...")
	pe := governance.NewPolicyEngine("policy.json")
	gm := governance.NewGovernanceManager(pe, nm)
	gm.Start(5 * time.Second) // Check every 5s
	// --------------------------------------------------

	var wg sync.WaitGroup

	// Start standard services
	standardServices := []struct {
		Name  string
		Start func(*network.NetworkManager, *governance.GovernanceManager)
	}{
		{"Core Server", server.Start},
		{"DNS Authority", dns.Start},
		{"Admin Panel", admin.Start},
	}

	for _, s := range standardServices {
		wg.Add(1)
		go func(name string, startFunc func(*network.NetworkManager, *governance.GovernanceManager)) {
			defer wg.Done()
			utils.LogInfo("Matrix", "Launching "+name+"...")
			startFunc(nm, gm)
		}(s.Name, s.Start)
	}

	// Start Services with Network Intelligence
	wg.Add(1)
	go func() {
		defer wg.Done()
		utils.LogInfo("Matrix", "Launching Network Gateway...")
		gateway.Start(nm, gm)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		utils.LogInfo("Matrix", "Launching Quantum Chat...")
		chat.Start(nm, gm)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		utils.LogInfo("Matrix", "Launching Digital Storage...")
		storage.Start(nm, gm)
	}()

	// Start Dashboard with Network Manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		utils.LogInfo("Matrix", "Launching System Dashboard...")
		dashboard.Start(nm, gm)
	}()

	utils.LogSuccess("System", "ALL SERVICES OPERATIONAL")

	fmt.Printf("\n  +--------------------------------------------------+\n")
	fmt.Printf("  |  🌐 UNIFIED HUB: http://%s:7000        |\n", localIP)
	fmt.Printf("  |  🚪 GATEWAY:     http://%s:8000        |\n", localIP)
	fmt.Printf("  +--------------------------------------------------+\n\n")

	// Wait for interrupt
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	utils.LogInfo("System", "Nexa Intelligence Layer Active. Press Ctrl+C to shutdown.")
	<-stop

	fmt.Println("\n[SYSTEM] Graceful shutdown initiated...")
	os.Exit(0)
}
