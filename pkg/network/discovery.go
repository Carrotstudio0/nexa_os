package network

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// DiscoveryBeacon represents a discovery beacon broadcast
type DiscoveryBeacon struct {
	DeviceID             string           `json:"device_id"`
	DeviceName           string           `json:"device_name"`
	Role                 DeviceRole       `json:"role"`
	MAC                  string           `json:"mac"`
	IPAddress            string           `json:"ip_address"`
	Port                 int              `json:"port"`
	SupportedConnections []ConnectionType `json:"supported_connections"`
	Timestamp            int64            `json:"timestamp"`
}

// DiscoveryResponse represents a response to a discovery beacon
type DiscoveryResponse struct {
	DeviceID             string           `json:"device_id"`
	DeviceName           string           `json:"device_name"`
	Role                 DeviceRole       `json:"role"`
	MAC                  string           `json:"mac"`
	IPAddress            string           `json:"ip_address"`
	Port                 int              `json:"port"`
	SupportedConnections []ConnectionType `json:"supported_connections"`
	SignalStrength       int              `json:"signal_strength"`
	Timestamp            int64            `json:"timestamp"`
}

// DeviceDiscovery handles device discovery
type DeviceDiscovery struct {
	mu                sync.RWMutex
	discoveredDevices map[string]*DiscoveryResponse
	broadcastAddr     net.UDPAddr
	listenAddr        net.UDPAddr
	conn              *net.UDPConn
	isRunning         bool
	stopChan          chan bool
	onDiscovered      func(*DiscoveryResponse)
}

// NewDeviceDiscovery creates a new device discovery instance
func NewDeviceDiscovery(broadcastIP string, port int) *DeviceDiscovery {
	return &DeviceDiscovery{
		discoveredDevices: make(map[string]*DiscoveryResponse),
		broadcastAddr: net.UDPAddr{
			Port: port,
			IP:   net.ParseIP(broadcastIP),
		},
		listenAddr: net.UDPAddr{
			Port: port,
			IP:   net.ParseIP("0.0.0.0"),
		},
		isRunning: false,
		stopChan:  make(chan bool, 1),
	}
}

// Start starts the discovery listener
func (dd *DeviceDiscovery) Start() error {
	dd.mu.Lock()
	if dd.isRunning {
		dd.mu.Unlock()
		return fmt.Errorf("discovery already running")
	}

	conn, err := net.ListenUDP("udp", &dd.listenAddr)
	if err != nil {
		dd.mu.Unlock()
		return fmt.Errorf("failed to start UDP listener: %v", err)
	}

	dd.conn = conn
	dd.isRunning = true
	dd.mu.Unlock()

	go dd.listenLoop()
	return nil
}

// Stop stops the discovery listener
func (dd *DeviceDiscovery) Stop() {
	dd.mu.Lock()
	if !dd.isRunning {
		dd.mu.Unlock()
		return
	}

	dd.isRunning = false
	dd.mu.Unlock()

	dd.stopChan <- true

	dd.mu.Lock()
	if dd.conn != nil {
		dd.conn.Close()
		dd.conn = nil
	}
	dd.mu.Unlock()
}

// Broadcast sends a discovery beacon
func (dd *DeviceDiscovery) Broadcast(beacon *DiscoveryBeacon) error {
	dd.mu.RLock()
	conn := dd.conn
	broadcastAddr := dd.broadcastAddr
	dd.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("discovery not running")
	}

	data, err := json.Marshal(beacon)
	if err != nil {
		return fmt.Errorf("failed to marshal beacon: %v", err)
	}

	_, err = conn.WriteToUDP(data, &broadcastAddr)
	return err
}

// listenLoop listens for discovery beacons
func (dd *DeviceDiscovery) listenLoop() {
	buffer := make([]byte, 4096)

	for {
		select {
		case <-dd.stopChan:
			return
		default:
			// Set read deadline to allow periodic shutdown check
			dd.conn.SetReadDeadline(time.Now().Add(5 * time.Second))

			n, remoteAddr, err := dd.conn.ReadFromUDP(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				if !dd.isRunning {
					return
				}
				continue
			}

			var beacon DiscoveryBeacon
			if err := json.Unmarshal(buffer[:n], &beacon); err != nil {
				continue
			}

			response := &DiscoveryResponse{
				DeviceID:             beacon.DeviceID,
				DeviceName:           beacon.DeviceName,
				Role:                 beacon.Role,
				MAC:                  beacon.MAC,
				IPAddress:            beacon.IPAddress,
				Port:                 beacon.Port,
				SupportedConnections: beacon.SupportedConnections,
				Timestamp:            time.Now().Unix(),
			}

			// Calculate signal strength based on network conditions (simplified)
			response.SignalStrength = dd.calculateSignalStrength(remoteAddr)

			dd.mu.Lock()
			dd.discoveredDevices[beacon.DeviceID] = response
			dd.mu.Unlock()

			if dd.onDiscovered != nil {
				dd.onDiscovered(response)
			}
		}
	}
}

// GetDiscoveredDevices returns all discovered devices
func (dd *DeviceDiscovery) GetDiscoveredDevices() map[string]*DiscoveryResponse {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	devices := make(map[string]*DiscoveryResponse)
	for id, dev := range dd.discoveredDevices {
		devices[id] = dev
	}

	return devices
}

// GetDiscoveredDevice returns a specific discovered device
func (dd *DeviceDiscovery) GetDiscoveredDevice(deviceID string) *DiscoveryResponse {
	dd.mu.RLock()
	defer dd.mu.RUnlock()

	return dd.discoveredDevices[deviceID]
}

// ClearDiscoveredDevices clears all discovered devices
func (dd *DeviceDiscovery) ClearDiscoveredDevices() {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	dd.discoveredDevices = make(map[string]*DiscoveryResponse)
}

// SetOnDiscovered sets the callback for device discovery
func (dd *DeviceDiscovery) SetOnDiscovered(callback func(*DiscoveryResponse)) {
	dd.mu.Lock()
	defer dd.mu.Unlock()

	dd.onDiscovered = callback
}

// calculateSignalStrength calculates signal strength based on network conditions
func (dd *DeviceDiscovery) calculateSignalStrength(remoteAddr *net.UDPAddr) int {
	// Simplified calculation: in a real implementation, this would measure
	// actual signal strength, packet loss, etc.
	// For now, return a default value
	return 75
}

// ProximityChecker checks proximity between devices
type ProximityChecker struct {
	mu              sync.RWMutex
	discovery       *DeviceDiscovery
	proximityRadius int // in meters (estimated)
	checkInterval   time.Duration
	isRunning       bool
	stopChan        chan bool
}

// NewProximityChecker creates a new proximity checker
func NewProximityChecker(discovery *DeviceDiscovery, radiusMeters int) *ProximityChecker {
	return &ProximityChecker{
		discovery:       discovery,
		proximityRadius: radiusMeters,
		checkInterval:   10 * time.Second,
		isRunning:       false,
		stopChan:        make(chan bool, 1),
	}
}

// Start starts the proximity checker
func (pc *ProximityChecker) Start() error {
	pc.mu.Lock()
	if pc.isRunning {
		pc.mu.Unlock()
		return fmt.Errorf("proximity checker already running")
	}
	pc.isRunning = true
	pc.mu.Unlock()

	go pc.checkLoop()
	return nil
}

// Stop stops the proximity checker
func (pc *ProximityChecker) Stop() {
	pc.mu.Lock()
	if !pc.isRunning {
		pc.mu.Unlock()
		return
	}
	pc.isRunning = false
	pc.mu.Unlock()

	pc.stopChan <- true
}

// checkLoop periodically checks device proximity
func (pc *ProximityChecker) checkLoop() {
	ticker := time.NewTicker(pc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pc.stopChan:
			return
		case <-ticker.C:
			pc.checkProximity()
		}
	}
}

// checkProximity checks proximity between discovered devices
func (pc *ProximityChecker) checkProximity() {
	devices := pc.discovery.GetDiscoveredDevices()

	for deviceID, device := range devices {
		// In a real implementation, this would use RSSI (Received Signal Strength Indicator)
		// to calculate actual proximity
		// For now, we use signal strength as a proxy
		estimatedDistance := pc.estimateDistance(device.SignalStrength)

		if estimatedDistance <= pc.proximityRadius {
			// Device is within proximity range
			_ = deviceID // Device is in range
		}
	}
}

// estimateDistance estimates distance based on signal strength
func (pc *ProximityChecker) estimateDistance(signalStrength int) int {
	// Simplified estimation: RSSI to distance conversion
	// In reality, this would be more complex and depend on various factors
	if signalStrength >= 80 {
		return 5 // meters
	} else if signalStrength >= 60 {
		return 20
	} else if signalStrength >= 40 {
		return 50
	}
	return 100
}
