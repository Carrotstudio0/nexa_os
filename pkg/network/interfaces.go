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
	if runtime.GOOS == "windows" {
		return enableWindowsHotspot(ssid, password)
	} else if runtime.GOOS == "linux" {
		return enableLinuxHotspot(ssid, password)
	}
	return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
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

// enableLinuxHotspot uses hostapd for Linux hotspot
func enableLinuxHotspot(ssid, password string) error {
	// This would require hostapd and dnsmasq to be installed
	cmd := exec.Command("hostapd", "-D", "nl80211", "-i", "wlan0", "-c", "/etc/hostapd/hostapd.conf", "-B")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start hostapd: %w", err)
	}
	return nil
}

// DisableHotspot disables the Wi-Fi hotspot
func DisableHotspot() error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("netsh", "wlan", "stop", "hostednetwork")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to stop hosted network: %w", err)
		}
	}
	return nil
}

// GetConnectedDevices returns a list of devices connected to the network
func GetConnectedDevices() ([]DeviceInfo, error) {
	if runtime.GOOS == "windows" {
		return getWindowsDevices()
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
