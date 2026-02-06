package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// NetworkInterface represents a network interface with details
type NetworkInterface struct {
	Name      string   `json:"name"`
	IP        string   `json:"ip"`
	MAC       string   `json:"mac"`
	IsActive  bool     `json:"is_active"`
	Network   string   `json:"network"`
	Addresses []string `json:"addresses"`
}

// GetLocalIP returns the main local IP address
func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// GetAllNetworkInterfaces returns all available network interfaces
func GetAllNetworkInterfaces() []NetworkInterface {
	var interfaces []NetworkInterface

	ifaces, err := net.Interfaces()
	if err != nil {
		return interfaces
	}

	for _, iface := range ifaces {
		ni := NetworkInterface{
			Name:     iface.Name,
			MAC:      iface.HardwareAddr.String(),
			IsActive: iface.Flags&net.FlagUp != 0,
		}

		addrs, err := iface.Addrs()
		if err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					ni.IP = ipnet.IP.String()
					ni.Network = ipnet.String()
					ni.Addresses = append(ni.Addresses, ipnet.IP.String())
				}
			}
		}

		interfaces = append(interfaces, ni)
	}

	return interfaces
}

// EnableHotspot enables Wi-Fi hotspot on the system
func EnableHotspot(ssid, password string) error {
	switch runtime.GOOS {
	case "windows":
		return enableWindowsHotspot(ssid, password)
	case "linux":
		return enableLinuxHotspot(ssid, password)
	case "darwin":
		return enableMacHotspot(ssid, password)
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// enableWindowsHotspot uses netsh to enable hotspot on Windows
func enableWindowsHotspot(ssid, password string) error {
	// First, set up the hosted network
	cmd := exec.Command("netsh", "wlan", "set", "hostednetwork", "mode=allow", fmt.Sprintf("ssid=%s", ssid), fmt.Sprintf("key=%s", password))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set hosted network: %w", err)
	}

	// Start the hosted network
	cmd = exec.Command("netsh", "wlan", "start", "hostednetwork")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start hosted network: %w", err)
	}

	return nil
}

// enableLinuxHotspot uses hostapd and dnsmasq for Linux hotspot
func enableLinuxHotspot(ssid, password string) error {
	// Check if hostapd is installed
	if _, err := exec.LookPath("hostapd"); err != nil {
		return fmt.Errorf("hostapd not found - install it with: apt-get install hostapd dnsmasq")
	}

	// For Linux, we'll use nmcli if available (NetworkManager)
	if _, err := exec.LookPath("nmcli"); err == nil {
		cmd := exec.Command("nmcli", "device", "wifi", "hotspot", "on", fmt.Sprintf("ssid=%s", ssid), fmt.Sprintf("password=%s", password))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to start hotspot via nmcli: %w", err)
		}
		return nil
	}

	// Fallback to hostapd
	cmd := exec.Command("hostapd", "-D", "nl80211", "-i", "wlan0", "-c", "/etc/hostapd/hostapd.conf", "-B")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start hostapd: %w", err)
	}
	return nil
}

// enableMacHotspot uses macOS native command for hotspot
func enableMacHotspot(ssid, password string) error {
	// macOS doesn't support programmatic WiFi hotspot easily
	// We'll document this as a limitation
	return fmt.Errorf("WiFi hotspot setup on macOS requires manual configuration.\n" +
		"Please use System Preferences > Sharing > Internet Sharing.\n" +
		"Or use: networksetup commands with proper permissions")
}

// DisableHotspot disables the Wi-Fi hotspot
func DisableHotspot() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("netsh", "wlan", "stop", "hostednetwork")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop hosted network: %w", err)
		}
		return nil
	case "linux":
		// Try nmcli first
		if _, err := exec.LookPath("nmcli"); err == nil {
			cmd := exec.Command("nmcli", "device", "wifi", "hotspot", "off")
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to stop hotspot via nmcli: %w", err)
			}
			return nil
		}
		// Fallback to pkill
		cmd := exec.Command("pkill", "hostapd")
		cmd.Run() // Ignore error
		return nil
	case "darwin":
		return fmt.Errorf("WiFi hotspot on macOS requires manual shutdown via System Preferences")
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// GetConnectedDevices returns a list of devices connected to the network
func GetConnectedDevices() ([]DeviceInfo, error) {
	if runtime.GOOS == "windows" {
		return getWindowsDevices()
	} else if runtime.GOOS == "linux" {
		return getLinuxDevices()
	} else if runtime.GOOS == "darwin" {
		return getMacDevices()
	}
	return nil, fmt.Errorf("device discovery not implemented for %s", runtime.GOOS)
}

// DeviceInfo represents information about a connected device
type DeviceInfo struct {
	Hostname  string    `json:"hostname"`
	IP        string    `json:"ip"`
	MAC       string    `json:"mac"`
	LastSeen  time.Time `json:"last_seen"`
	Connected bool      `json:"connected"`
}

func getWindowsDevices() ([]DeviceInfo, error) {
	var devices []DeviceInfo

	// Use arp -a to list devices
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		return devices, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && strings.Contains(parts[0], ".") {
			device := DeviceInfo{
				IP:        parts[0],
				MAC:       parts[1],
				LastSeen:  time.Now(),
				Connected: true,
			}
			devices = append(devices, device)
		}
	}

	return devices, nil
}

// getLinuxDevices retrieves connected devices on Linux using arp-scan or arp
func getLinuxDevices() ([]DeviceInfo, error) {
	var devices []DeviceInfo

	// Try using 'arp' command first
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: return empty list instead of error
		return devices, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			devices = append(devices, DeviceInfo{
				Hostname: fields[0],
				IP:       strings.Trim(fields[1], "()"),
				MAC:      fields[2],
			})
		}
	}

	return devices, nil
}

// getMacDevices retrieves connected devices on macOS using arp
func getMacDevices() ([]DeviceInfo, error) {
	var devices []DeviceInfo

	// macOS uses 'arp -a' similar to Linux
	cmd := exec.Command("arp", "-a")
	output, err := cmd.Output()
	if err != nil {
		return devices, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "at") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				devices = append(devices, DeviceInfo{
					Hostname: fields[0],
					IP:       strings.Trim(fields[1], "()"),
					MAC:      fields[3],
				})
			}
		}
	}

	return devices, nil
}

// NetworkStats provides network statistics
type NetworkStats struct {
	TotalConnections int          `json:"total_connections"`
	ActiveInterfaces []string     `json:"active_interfaces"`
	Devices          []DeviceInfo `json:"devices"`
	LocalIP          string       `json:"local_ip"`
	Timestamp        time.Time    `json:"timestamp"`
}

// GetNetworkStats returns comprehensive network statistics
func GetNetworkStats() NetworkStats {
	devices, _ := GetConnectedDevices()
	interfaces := GetAllNetworkInterfaces()

	var activeIfaces []string
	for _, iface := range interfaces {
		if iface.IsActive && iface.IP != "" {
			activeIfaces = append(activeIfaces, iface.Name)
		}
	}

	return NetworkStats{
		TotalConnections: len(devices),
		ActiveInterfaces: activeIfaces,
		Devices:          devices,
		LocalIP:          GetLocalIP(),
		Timestamp:        time.Now(),
	}
}

// JSONStats returns stats as JSON
func (ns NetworkStats) JSONStats() []byte {
	data, _ := json.MarshalIndent(ns, "", "  ")
	return data
}
