//go:build windows

package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// UpdateHostsFile adds or updates a local domain mapping in the Windows hosts file
func UpdateHostsFile(domain string, ip string) error {
	hostsPath := `C:\Windows\System32\drivers\etc\hosts`
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		// Log but don't fail startup
		LogWarning("System", fmt.Sprintf("Could not read hosts file: %v", err))
		return nil
	}

	contentStr := string(content)
	line := fmt.Sprintf("%s %s # NEXA_AUTO_DOMAIN", ip, domain)

	if !strings.Contains(contentStr, domain) {
		// Add the domain entry
		f, err := os.OpenFile(hostsPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			LogWarning("System", fmt.Sprintf("Could not update hosts file: %v", err))
			return nil
		}
		defer f.Close()
		_, err = f.WriteString("\r\n" + line)
		if err != nil {
			LogWarning("System", fmt.Sprintf("Could not write to hosts file: %v", err))
			return nil
		}
		LogSuccess("System", fmt.Sprintf("Added '%s' to hosts file", domain))
	}
	return nil
}

// SetupFirewallRules opens necessary ports in Windows Firewall for professional network access
func SetupFirewallRules() {
	// These commands are executed with error suppression to avoid blocking startup
	ports := []struct {
		Name     string
		Protocol string
		Port     string
	}{
		{"NEXA DNS PRO", "UDP", "53"},
		{"NEXA WEB PRO", "TCP", "80"},
		{"NEXA GATEWAY", "TCP", "8000"},
		{"NEXA Dashboard", "TCP", "7000"},
		{"NEXA Admin", "TCP", "8080"},
		{"NEXA Storage", "TCP", "8081"},
		{"NEXA Chat", "TCP", "8082"},
		{"NEXA Core", "TCP", "1413"},
		{"NEXA mDNS", "UDP", "5353"},
	}

	for _, p := range ports {
		// Delete rule if exists, then add it
		delCmd := fmt.Sprintf(`netsh advfirewall firewall delete rule name="%s" >nul 2>&1`, p.Name)
		addCmd := fmt.Sprintf(`netsh advfirewall firewall add rule name="%s" dir=in action=allow protocol=%s localport=%s profile=any >nul 2>&1`, p.Name, p.Protocol, p.Port)

		exec.Command("cmd", "/c", delCmd).Run()
		exec.Command("cmd", "/c", addCmd).Run()
	}

	// Try to stop Windows hidden HTTP service if it's blocking port 80
	exec.Command("powershell", "-Command", "Stop-Service -Name W3SVC -ErrorAction SilentlyContinue").Run()
}

// StartHotspot runs the generic hotspot script
func StartHotspot() {
	// Matrix Wireless Matrix Pulse
	exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-File", "./scripts/enable-hotspot.ps1").Start()
}

// OpenURL opens a URL in the default browser
func OpenURL(url string) {
	exec.Command("cmd", "/c", "start", url).Start()
}

// CreateShortcuts creates desktop shortcuts (Windows Specific)
func CreateShortcuts(execPath string) {
	desktop, _ := os.UserHomeDir()
	desktop = filepath.Join(desktop, "Desktop")

	createShortcut(filepath.Join(desktop, "Nexa.lnk"), execPath, "Nexa Operating System")
	createShortcut(filepath.Join(desktop, "Nexa Dashboard.lnk"), "http://localhost:7000", "Nexa Control Center")
}

func createShortcut(shortcutPath, targetPath, description string) {
	psCmd := fmt.Sprintf("$s=(New-Object -COM WScript.Shell).CreateShortcut('%s');$s.TargetPath='%s';$s.Description='%s';$s.Save()",
		shortcutPath, targetPath, description)
	exec.Command("powershell", "-Command", psCmd).Run()
}
