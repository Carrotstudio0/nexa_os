//go:build !windows

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// UpdateHostsFile is a stub for non-Windows systems
func UpdateHostsFile(domain string, ip string) error {
	LogWarning("System", "Hosts file update skipped (Requires Root/Sudo on Linux/Mac). Please add '"+ip+" "+domain+"' to /etc/hosts manually.")
	return nil
}

// SetupFirewallRules is a stub for non-Windows systems
func SetupFirewallRules() {
	LogInfo("System", "Skipping Windows Firewall setup (Not on Windows). Ensure ports 80, 53, 7000-8082 are open.")
}

// StartHotspot is a stub
func StartHotspot() {
	LogInfo("System", "Hotspot automation is currently Windows-only.")
}

// OpenURL opens a URL in the default browser using portable commands
func OpenURL(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	default:
		return
	}
	exec.Command(cmd, args...).Start()
}

// CreateShortcuts creates desktop shortcuts (Linux-specific)
// Creates .desktop files for easier application launching on Linux systems
func CreateShortcuts(execPath string) {
	if runtime.GOOS != "linux" {
		return
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		return
	}

	desktopDir := filepath.Join(homeDir, ".local/share/applications")
	os.MkdirAll(desktopDir, 0755)

	// Create Nexa launcher .desktop file
	desktopFile := filepath.Join(desktopDir, "nexa.desktop")
	content := `[Desktop Entry]
Version=1.0
Type=Application
Name=Nexa Ultimate
Exec=` + execPath + `
Icon=network
Categories=Utility;
Terminal=false
`
	os.WriteFile(desktopFile, []byte(content), 0644)
}
