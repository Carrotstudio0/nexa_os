package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MultiX0/nexa/pkg/analytics"
	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/services/admin"
	"github.com/MultiX0/nexa/pkg/services/chat"
	"github.com/MultiX0/nexa/pkg/services/dashboard"
	"github.com/MultiX0/nexa/pkg/services/dns"
	"github.com/MultiX0/nexa/pkg/services/gateway"
	"github.com/MultiX0/nexa/pkg/services/server"
	"github.com/MultiX0/nexa/pkg/services/storage"
	"github.com/MultiX0/nexa/pkg/services/web"
	"github.com/MultiX0/nexa/pkg/utils"
)

type Service struct {
	Name    string
	Running bool
	Cancel  context.CancelFunc
	Start   func(*network.NetworkManager, *governance.GovernanceManager)
}

func main() {
	// Initialize Configuration
	cfg, err := config.Load()
	if err != nil {
		utils.LogWarning("Config", fmt.Sprintf("Loader issue: %v", err))
		// We continue because Load() sets defaults even on error
	}

	// MATRIX PRO: Auto-deploy Wireless Matrix (Hotspot)
	utils.LogInfo("Nucleus", "Deploying Wireless Matrix Pulse...")
	go func() {
		// Run hotspot via cross-platform wrapper
		utils.StartHotspot()
	}()

	// MATRIX PRO: Secure Gateway & Firewall Orchestration
	utils.LogInfo("Security", "Configuring network access controls...")
	utils.SetupFirewallRules()

	// Print Professional Mobile Connectivity Banner
	gatewayURL := fmt.Sprintf("http://%s:%d", utils.GetLocalIP(), cfg.Services.Gateway.Port)
	fmt.Printf("\n   \033[36m+--------------------------------------------------------------+\033[0m\n")
	fmt.Printf("   \033[36m| \033[37m\033[1m  MOBILE ACCESS:   %-32s \033[0m\033[36m          |\033[0m\n", gatewayURL)
	fmt.Printf("   \033[36m| \033[37m\033[1m  PROFESSIONAL:    http://hub.n                             \033[0m\033[36m|\033[0m\n")
	fmt.Printf("   \033[36m+--------------------------------------------------------------+\033[0m\n")
	fmt.Println()

	utils.UpdateHostsFile("demo.n", "127.0.0.1")
	utils.UpdateHostsFile("admin.n", "127.0.0.1")

	// Silence noise for professional output
	os.Setenv("PORT_SERVER", "1413")
	os.Setenv("PORT_DNS", "1112")
	os.Setenv("PORT_WEB", "8080")

	utils.PrintBanner(cfg.System.Name, cfg.System.Version)

	ctx, cancel := context.WithCancel(context.Background())
	_ = ctx
	defer cancel()

	// --- System Infrastructure ---
	utils.LogInfo("Nucleus", "Hypervising System Matrix...")
	localIP := utils.GetLocalIP()

	nm := network.NewNetworkManager(network.ConnectionConfig{
		Timeout:           10 * time.Second,
		MaxRetries:        5,
		HeartbeatInterval: 15 * time.Second,
		ConnectionType:    network.ConnectionWiFi,
	})

	// Register Core Nodes
	base, err := nm.RegisterPrimaryBase("base-core", "Nexa Nucleus Core", "00:00:00:00:00:00", localIP, cfg.Services.Dashboard.Port)
	if err == nil {
		base.IsOnline = true
		base.UpdateOnlineStatus(true)
	}

	gm := governance.NewGovernanceManager(governance.NewPolicyEngine("policy.json"), nm)
	gm.Start(5 * time.Second)

	// MATRIX PRO: Initialize Global Analytics Matrix
	analytics.GetManager().SetGovernance(gm)

	// --- Integrated Module Matrix ---
	services := []struct {
		ID, Name string
		Port     int
		StartFn  func(*network.NetworkManager, *governance.GovernanceManager)
	}{
		{"svc-dns", "DNS Authority", cfg.Services.DNS.Port, dns.Start},
		{"svc-gateway", "Matrix Gateway", cfg.Services.Gateway.Port, gateway.Start},
		{"svc-admin", "Admin Center", cfg.Services.Admin.Port, admin.Start},
		{"svc-storage", "Digital Vault", cfg.Services.Storage.Port, storage.Start},
		{"svc-chat", "Matrix Chat", cfg.Services.Chat.Port, chat.Start},
		{"svc-dashboard", "Intelligence Hub", cfg.Services.Dashboard.Port, dashboard.Start},
		{"svc-web", "Web Elite", cfg.Services.Web.Port, web.Start},
		{"svc-core", "Core Nucleus", cfg.Server.Port, server.Start},
	}

	utils.LogInfo("Matrix", fmt.Sprintf("Synchronizing %d integrated modules...", len(services)))

	for _, s := range services {
		node, err := nm.RegisterDevice(s.ID, s.Name, "", localIP, s.Port, network.RoleGateway)
		if err == nil {
			node.IsOnline = true
			node.UpdateOnlineStatus(true)
			nm.CreateConnection(base.ID, node.ID, network.ConnectionMesh)
		}

		go func(name string, startFn func(*network.NetworkManager, *governance.GovernanceManager)) {
			defer func() {
				if r := recover(); r != nil {
					utils.LogError("Supervisor", "Critical failure in "+name, fmt.Errorf("%v", r))
				}
			}()
			startFn(nm, gm)
		}(s.Name, s.StartFn)

		utils.LogSuccess("Module", "Synchronized: "+s.Name)
		time.Sleep(150 * time.Millisecond)
	}

	dashboardURL := fmt.Sprintf("http://%s:%d", localIP, cfg.Services.Dashboard.Port)
	utils.LogSuccess("System", "Matrix is fully operational.")

	fmt.Printf("\n  %s╔══════════════════════════════════════════════════════════════╗%s\n", utils.ColorCyan, utils.ColorReset)
	fmt.Printf("  %s║%s  🌐 INTELLIGENCE HUB:  %s        %s║%s\n", utils.ColorCyan, utils.ColorWhite+utils.ColorBold, dashboardURL, utils.ColorReset, utils.ColorCyan)
	fmt.Printf("  %s║%s  🚪 MATRIX GATEWAY:    http://%s:%d        %s║%s\n", utils.ColorCyan, utils.ColorWhite+utils.ColorBold, localIP, cfg.Services.Gateway.Port, utils.ColorReset, utils.ColorCyan)
	fmt.Printf("  %s╚══════════════════════════════════════════════════════════════╝%s\n\n", utils.ColorCyan, utils.ColorReset)

	// PRO TOUCH: Automate Intelligence Hub launch
	utils.LogInfo("Sync", "Deploying UI to local workspace...")
	go func() {
		time.Sleep(2 * time.Second)
		// Using localhost for guaranteed local access regardless of IP stability
		localDashboard := fmt.Sprintf("http://localhost:%d", cfg.Services.Dashboard.Port)
		utils.OpenURL(localDashboard)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	utils.LogWarning("System", "De-synchronizing Matrix logic...")
	fmt.Println("\n[SYSTEM] Nexa Ultimate offline.")
	os.Exit(0)
}
