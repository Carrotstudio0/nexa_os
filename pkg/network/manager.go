package network

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// NetworkManager manages all devices and connections in the network
type NetworkManager struct {
	mu                   sync.RWMutex
	topology             *NetworkTopology
	handlers             map[string]*ConnectionHandler
	config               ConnectionConfig
	onDeviceConnected    func(*Device)
	onDeviceDisconnected func(*Device)
	stopChan             chan bool
	monitoringActive     bool
}

// NewNetworkManager creates a new network manager
func NewNetworkManager(config ConnectionConfig) *NetworkManager {
	return &NetworkManager{
		topology:         NewNetworkTopology(),
		handlers:         make(map[string]*ConnectionHandler),
		config:           config,
		stopChan:         make(chan bool),
		monitoringActive: false,
	}
}

// RegisterPrimaryBase registers the primary base station
func (nm *NetworkManager) RegisterPrimaryBase(id, name, mac, ip string, port int) (*Device, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.topology.PrimaryBase != nil {
		return nil, fmt.Errorf("primary base already registered")
	}

	primaryBase := NewDevice(id, name, RolePrimaryBase, mac, ip, port)
	nm.topology.PrimaryBase = primaryBase
	nm.topology.AddDevice(primaryBase)

	return primaryBase, nil
}

// RegisterDevice registers a new device in the network
func (nm *NetworkManager) RegisterDevice(id, name, mac, ip string, port int, role DeviceRole) (*Device, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.topology.Devices[id]; exists {
		return nil, fmt.Errorf("device %s already registered", id)
	}

	device := NewDevice(id, name, role, mac, ip, port)
	nm.topology.AddDevice(device)

	return device, nil
}

// ConnectDevice establishes a connection to a device
func (nm *NetworkManager) ConnectDevice(deviceID string, connType ConnectionType) error {
	nm.mu.Lock()
	device, exists := nm.topology.Devices[deviceID]
	nm.mu.Unlock()

	if !exists {
		return fmt.Errorf("device %s not found", deviceID)
	}

	device.ConnectionType = connType

	// Create connection handler
	handler := NewConnectionHandler(device, nm.config)
	if err := handler.Connect(); err != nil {
		return fmt.Errorf("failed to connect device %s: %v", deviceID, err)
	}

	nm.mu.Lock()
	nm.handlers[deviceID] = handler
	nm.mu.Unlock()

	// Start heartbeat
	handler.StartHeartbeat()

	// Trigger callback
	if nm.onDeviceConnected != nil {
		nm.onDeviceConnected(device)
	}

	return nil
}

// DisconnectDevice closes the connection to a device
func (nm *NetworkManager) DisconnectDevice(deviceID string) error {
	nm.mu.Lock()
	handler, exists := nm.handlers[deviceID]
	nm.mu.Unlock()

	if !exists {
		return fmt.Errorf("device %s not connected", deviceID)
	}

	device := nm.topology.GetDevice(deviceID)
	if device != nil {
		device.UpdateOnlineStatus(false)
	}

	err := handler.Disconnect()

	nm.mu.Lock()
	delete(nm.handlers, deviceID)
	nm.mu.Unlock()

	// Trigger callback
	if nm.onDeviceDisconnected != nil && device != nil {
		nm.onDeviceDisconnected(device)
	}

	return err
}

// CreateConnection establishes a logical connection between two devices
func (nm *NetworkManager) CreateConnection(sourceID, targetID string, connType ConnectionType) (*DeviceConnection, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	_, sourceExists := nm.topology.Devices[sourceID]
	_, targetExists := nm.topology.Devices[targetID]

	if !sourceExists || !targetExists {
		return nil, fmt.Errorf("source or target device not found")
	}

	connID := nm.generateConnectionID(sourceID, targetID)
	conn := NewDeviceConnection(sourceID, targetID, connType)
	conn.ID = connID
	conn.IsActive = true

	nm.topology.AddConnection(conn)

	return conn, nil
}

// RemoveConnection removes a connection between two devices
func (nm *NetworkManager) RemoveConnection(connID string) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if _, exists := nm.topology.Connections[connID]; !exists {
		return fmt.Errorf("connection %s not found", connID)
	}

	nm.topology.RemoveConnection(connID)
	return nil
}

// SendCommandToDevice sends a command to a specific device
func (nm *NetworkManager) SendCommandToDevice(deviceID string, command string, args map[string]interface{}) error {
	nm.mu.RLock()
	handler, exists := nm.handlers[deviceID]
	nm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("device %s not connected", deviceID)
	}

	msg := CommandMessage{
		Type:      "command",
		DeviceID:  deviceID,
		Timestamp: time.Now().Unix(),
		Command:   command,
		Args:      args,
	}

	return handler.SendMessage(msg)
}

// BroadcastMessage sends a message to all connected devices
func (nm *NetworkManager) BroadcastMessage(message interface{}) error {
	nm.mu.RLock()
	handlers := make(map[string]*ConnectionHandler)
	for id, handler := range nm.handlers {
		handlers[id] = handler
	}
	nm.mu.RUnlock()

	var lastErr error
	for _, handler := range handlers {
		if err := handler.SendMessage(message); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GetDevice retrieves device information
func (nm *NetworkManager) GetDevice(deviceID string) *Device {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	return nm.topology.GetDevice(deviceID)
}

// GetTopology returns the current network topology
func (nm *NetworkManager) GetTopology() *NetworkTopology {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Create a copy to avoid external modifications
	topoCopy := &NetworkTopology{
		PrimaryBase: nm.topology.PrimaryBase,
		Devices:     make(map[string]*Device),
		Connections: make(map[string]*DeviceConnection),
		UpdatedAt:   nm.topology.UpdatedAt,
	}

	for id, device := range nm.topology.Devices {
		topoCopy.Devices[id] = device
	}

	for id, conn := range nm.topology.Connections {
		topoCopy.Connections[id] = conn
	}

	return topoCopy
}

// GetConnectedDevices returns a list of all connected devices
func (nm *NetworkManager) GetConnectedDevices() []*Device {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	var devices []*Device
	for _, device := range nm.topology.Devices {
		if device.IsOnline {
			devices = append(devices, device)
		}
	}

	return devices
}

// GetDevicesByRole returns all devices with a specific role
func (nm *NetworkManager) GetDevicesByRole(role DeviceRole) []*Device {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	var devices []*Device
	for _, device := range nm.topology.Devices {
		if device.Role == role {
			devices = append(devices, device)
		}
	}

	return devices
}

// SetOnDeviceConnected sets the callback for device connection
func (nm *NetworkManager) SetOnDeviceConnected(callback func(*Device)) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.onDeviceConnected = callback
}

// SetOnDeviceDisconnected sets the callback for device disconnection
func (nm *NetworkManager) SetOnDeviceDisconnected(callback func(*Device)) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.onDeviceDisconnected = callback
}

// StartMonitoring starts the network monitoring routine
func (nm *NetworkManager) StartMonitoring() {
	nm.mu.Lock()
	if nm.monitoringActive {
		nm.mu.Unlock()
		return
	}
	nm.monitoringActive = true
	nm.mu.Unlock()

	go nm.monitorLoop()
}

// StopMonitoring stops the network monitoring routine
func (nm *NetworkManager) StopMonitoring() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.monitoringActive {
		nm.stopChan <- true
		nm.monitoringActive = false
	}
}

// monitorLoop periodically checks device health
func (nm *NetworkManager) monitorLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-nm.stopChan:
			return
		case <-ticker.C:
			nm.checkDeviceHealth()
		}
	}
}

// checkDeviceHealth checks the health of all connected devices
func (nm *NetworkManager) checkDeviceHealth() {
	nm.mu.RLock()
	handlers := make(map[string]*ConnectionHandler)
	for id, handler := range nm.handlers {
		handlers[id] = handler
	}
	nm.mu.RUnlock()

	now := time.Now()
	timeout := 2 * time.Minute

	for deviceID, handler := range handlers {
		lastSeen := handler.GetLastMessageTime()

		if now.Sub(lastSeen) > timeout {
			// Device is considered offline
			device := nm.GetDevice(deviceID)
			if device != nil && device.IsOnline {
				nm.DisconnectDevice(deviceID)
			}
		}
	}
}

// generateConnectionID generates a unique connection ID
func (nm *NetworkManager) generateConnectionID(sourceID, targetID string) string {
	hash := md5.Sum([]byte(sourceID + "-" + targetID + "-" + time.Now().String()))
	return hex.EncodeToString(hash[:])
}

// GetNetworkStats returns network statistics
// Note: NetworkStats is defined in interfaces.go
func (nm *NetworkManager) GetNetworkStats() NetworkStats {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// Return basic network stats with current connection info
	stats := NetworkStats{
		TotalConnections: len(nm.topology.Connections),
		ActiveInterfaces: []string{},
		Devices:          []DeviceInfo{},
		Timestamp:        time.Now(),
	}

	return stats
}
