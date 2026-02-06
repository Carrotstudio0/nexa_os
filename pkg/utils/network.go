package utils

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

// GetLocalIP returns the primary local IP address, preferring 192.168.x.x
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	// First pass: look for 192.168 addresses (standard WiFi/Ethernet)
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			ip := ipnet.IP.To4()
			if ip != nil && ip[0] == 192 && ip[1] == 168 {
				return ip.String()
			}
		}
	}

	// Second pass: return any valid IPv4
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// GetMACAddress returns the first non-empty MAC address
func GetMACAddress() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "00:00:00:00:00:00"
	}
	for _, iface := range interfaces {
		if len(iface.HardwareAddr) > 0 {
			return iface.HardwareAddr.String()
		}
	}
	return "00:00:00:00:00:00"
}

// FormatSize formats bytes into a human readable string
func FormatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// UpdateHostsFile adds or updates a local domain mapping in the Windows hosts file
func UpdateHostsFile(domain string, ip string) error {
	hostsPath := `C:\Windows\System32\drivers\etc\hosts`
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		return err
	}

	line := fmt.Sprintf("\r\n%s %s # NEXA_AUTO_DOMAIN", ip, domain)
	if !strings.Contains(string(content), domain) {
		f, err := os.OpenFile(hostsPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(line)
		return err
	}
	return nil
}

// SetupFirewallRules opens necessary ports in Windows Firewall for professional network access
func SetupFirewallRules() {
	commands := []string{
		`netsh advfirewall firewall delete rule name="NEXA DNS PRO"`,
		`netsh advfirewall firewall delete rule name="NEXA WEB PRO"`,
		`netsh advfirewall firewall delete rule name="NEXA GATEWAY"`,
		`netsh advfirewall firewall add rule name="NEXA DNS PRO" dir=in action=allow protocol=UDP localport=53 profile=any`,
		`netsh advfirewall firewall add rule name="NEXA WEB PRO" dir=in action=allow protocol=TCP localport=80 profile=any`,
		`netsh advfirewall firewall add rule name="NEXA GATEWAY" dir=in action=allow protocol=TCP localport=8000 profile=any`,
		`netsh advfirewall firewall add rule name="NEXA mDNS" dir=in action=allow protocol=UDP localport=5353 profile=any`,
	}

	for _, cmd := range commands {
		exec.Command("powershell", "-Command", cmd).Run()
	}

	// Stop Windows hidden HTTP service if it's blocking port 80
	exec.Command("powershell", "-Command", "Stop-Service -Name W3SVC -ErrorAction SilentlyContinue").Run()
}
