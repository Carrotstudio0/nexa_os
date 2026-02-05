package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SetupSystem handles all the technical installation tasks for Windows
func SetupSystem() error {
	LogInfo("Setup", "Initializing Windows System Integration...")

	// 1. Create Data Directories
	execPath, _ := os.Executable()
	baseDir := filepath.Dir(execPath)
	dataDir := filepath.Join(baseDir, "data")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data dir: %v", err)
	}
	LogSuccess("Setup", "Data directories prepared.")

	// 2. Open Firewall Ports (Hub, Gateway, Admin, Chat, DNS)
	ports := []struct {
		Name string
		Port string
		Prot string
	}{
		{"Nexa Hub", "7000", "TCP"},
		{"Nexa Gateway", "8000", "TCP"},
		{"Nexa Admin", "8080", "TCP"},
		{"Nexa Storage", "8081", "TCP"},
		{"Nexa Chat", "8082", "TCP"},
		{"Nexa DNS", "53", "UDP"},
	}

	for _, p := range ports {
		cmd := exec.Command("netsh", "advfirewall", "firewall", "add", "rule",
			"name="+p.Name,
			"dir=in",
			"action=allow",
			"protocol="+p.Prot,
			"localport="+p.Port)
		if err := cmd.Run(); err != nil {
			LogWarning("Setup", fmt.Sprintf("Firewall rule exists or failed for %s: %v", p.Name, err))
		}
	}
	LogSuccess("Setup", "Network ports authorized in Firewall.")

	// 3. Create Shortcuts (Using PowerShell for maximum compatibility)
	desktop, _ := os.UserHomeDir()
	desktop = filepath.Join(desktop, "Desktop")

	createShortcut(filepath.Join(desktop, "Nexa.lnk"), execPath, "Nexa Operating System")
	createShortcut(filepath.Join(desktop, "Nexa Dashboard.lnk"), "http://localhost:7000", "Nexa Control Center")

	LogSuccess("Setup", "Desktop shortcuts established.")
	return nil
}

func createShortcut(shortcutPath, targetPath, description string) {
	// Simple PowerShell command to create a shortcut
	psCmd := fmt.Sprintf("$s=(New-Object -COM WScript.Shell).CreateShortcut('%s');$s.TargetPath='%s';$s.Description='%s';$s.Save()",
		shortcutPath, targetPath, description)

	exec.Command("powershell", "-Command", psCmd).Run()
}
