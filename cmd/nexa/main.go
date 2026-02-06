package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
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
	// Silence noise for professional output
	os.Setenv("PORT_SERVER", "1413")
	os.Setenv("PORT_DNS", "1112")
	os.Setenv("PORT_WEB", "8080")

	utils.PrintBanner("NEXA ULTIMATE", "v4.0.0-PRO")

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
	base, err := nm.RegisterPrimaryBase("base-core", "Nexa Nucleus Core", "00:00:00:00:00:00", localIP, 7000)
	if err == nil {
		base.IsOnline = true
		base.UpdateOnlineStatus(true)
	}

	gm := governance.NewGovernanceManager(governance.NewPolicyEngine("policy.json"), nm)
	gm.Start(5 * time.Second)

	// --- Integrated Module Matrix ---
	services := []struct {
		ID, Name string
		Port     int
		StartFn  func(*network.NetworkManager, *governance.GovernanceManager)
	}{
		{"svc-dns", "DNS Authority", 53, dns.Start},
		{"svc-gateway", "Matrix Gateway", 8000, gateway.Start},
		{"svc-admin", "Admin Center", 8080, admin.Start},
		{"svc-storage", "Digital Vault", 8081, storage.Start},
		{"svc-chat", "Matrix Chat", 8082, chat.Start},
		{"svc-dashboard", "Intelligence Hub", 7000, dashboard.Start},
		{"svc-web", "Web Elite", 3000, web.Start},
		{"svc-core", "Core Nucleus", 1413, server.Start},
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

	dashboardURL := fmt.Sprintf("http://%s:7000", localIP)
	utils.LogSuccess("System", "Matrix is fully operational.")

	fmt.Printf("\n  %s╔══════════════════════════════════════════════════════════════╗%s\n", utils.ColorCyan, utils.ColorReset)
	fmt.Printf("  %s║%s  🌐 INTELLIGENCE HUB:  %s        %s║%s\n", utils.ColorCyan, utils.ColorWhite+utils.ColorBold, dashboardURL, utils.ColorReset, utils.ColorCyan)
	fmt.Printf("  %s║%s  🚪 MATRIX GATEWAY:    http://%s:8000        %s║%s\n", utils.ColorCyan, utils.ColorWhite+utils.ColorBold, localIP, utils.ColorReset, utils.ColorCyan)
	fmt.Printf("  %s╚══════════════════════════════════════════════════════════════╝%s\n\n", utils.ColorCyan, utils.ColorReset)

	// PRO TOUCH: Automate Intelligence Hub launch
	utils.LogInfo("Sync", "Deploying UI to local workspace...")
	go func() {
		time.Sleep(2 * time.Second)
		// Using localhost for guaranteed local access regardless of IP stability
		localDashboard := "http://localhost:7000"
		exec.Command("cmd", "/c", "start", localDashboard).Start()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	utils.LogWarning("System", "De-synchronizing Matrix logic...")
	fmt.Println("\n[SYSTEM] Nexa Ultimate offline.")
	os.Exit(0)
}
