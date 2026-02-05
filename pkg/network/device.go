package network

import (
	"time"
)

// ConnectionType represents the type of connection method
type ConnectionType string

const (
	ConnectionWiFi       ConnectionType = "wifi"
	ConnectionBluetooth  ConnectionType = "bluetooth"
	ConnectionWiFiDirect ConnectionType = "wifi_direct"
	ConnectionMesh       ConnectionType = "mesh"
	ConnectionHotspot    ConnectionType = "hotspot"
)

// DeviceRole represents the role of a device in the network
type DeviceRole string

const (
	RolePrimaryBase DeviceRole = "primary_base"
	RoleGateway     DeviceRole = "gateway"
	RoleNode        DeviceRole = "node"
)

// Device represents a networked device
type Device struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Role           DeviceRole             `json:"role"`
	MAC            string                 `json:"mac"`
	IPAddress      string                 `json:"ip_address"`
	Port           int                    `json:"port"`
	ConnectionType ConnectionType         `json:"connection_type"`
	SignalStrength int                    `json:"signal_strength"` // -1 to 100
	IsOnline       bool                   `json:"is_online"`
	LastSeen       time.Time              `json:"last_seen"`
	RegisteredAt   time.Time              `json:"registered_at"`
	Metadata       map[string]interface{} `json:"metadata"`
	Metrics        DeviceMetrics          `json:"metrics"`
}

// DeviceMetrics represents performance metrics for a device
type DeviceMetrics struct {
	LatencyMS      int64                  `json:"latency_ms"`
	RequestsPerSec float64                `json:"requests_per_sec"`
	ErrorRate      float64                `json:"error_rate"`
	LastActivity   int64                  `json:"last_activity"`
	Custom         map[string]interface{} `json:"custom,omitempty"`
}

// DeviceConnection represents a connection between two devices
type DeviceConnection struct {
	ID             string         `json:"id"`
	SourceDeviceID string         `json:"source_device_id"`
	TargetDeviceID string         `json:"target_device_id"`
	ConnectionType ConnectionType `json:"connection_type"`
	IsActive       bool           `json:"is_active"`
	LatencyMS      int            `json:"latency_ms"`
	Bandwidth      int64          `json:"bandwidth"`  // in bytes
	ErrorRate      float32        `json:"error_rate"` // 0.0 to 1.0
	EstablishedAt  time.Time      `json:"established_at"`
	LastHeartbeat  time.Time      `json:"last_heartbeat"`
}

// NetworkTopology represents the current network structure
type NetworkTopology struct {
	PrimaryBase    *Device                           `json:"primary_base"`
	Devices        map[string]*Device                `json:"devices"`
	Connections    map[string]*DeviceConnection      `json:"connections"`
	ServiceMetrics map[string]map[string]interface{} `json:"service_metrics"`
	UpdatedAt      time.Time                         `json:"updated_at"`
}

// NewDevice creates a new device instance
func NewDevice(id, name string, role DeviceRole, mac, ip string, port int) *Device {
	return &Device{
		ID:             id,
		Name:           name,
		Role:           role,
		MAC:            mac,
		IPAddress:      ip,
		Port:           port,
		SignalStrength: -1,
		IsOnline:       false,
		RegisteredAt:   time.Now(),
		LastSeen:       time.Now(),
		Metadata:       make(map[string]interface{}),
	}
}

// NewDeviceConnection creates a new connection between devices
func NewDeviceConnection(sourceID, targetID string, connType ConnectionType) *DeviceConnection {
	return &DeviceConnection{
		SourceDeviceID: sourceID,
		TargetDeviceID: targetID,
		ConnectionType: connType,
		IsActive:       false,
		EstablishedAt:  time.Now(),
		LastHeartbeat:  time.Now(),
	}
}

// UpdateOnlineStatus updates the device's online status
func (d *Device) UpdateOnlineStatus(online bool) {
	d.IsOnline = online
	d.LastSeen = time.Now()
}

// UpdateSignalStrength updates the signal strength
func (d *Device) UpdateSignalStrength(strength int) {
	if strength < -1 {
		strength = -1
	}
	if strength > 100 {
		strength = 100
	}
	d.SignalStrength = strength
}

// NewNetworkTopology creates a new network topology
func NewNetworkTopology() *NetworkTopology {
	return &NetworkTopology{
		Devices:        make(map[string]*Device),
		Connections:    make(map[string]*DeviceConnection),
		ServiceMetrics: make(map[string]map[string]interface{}),
		UpdatedAt:      time.Now(),
	}
}

// AddDevice adds a device to the topology
func (t *NetworkTopology) AddDevice(device *Device) {
	if t.Devices == nil {
		t.Devices = make(map[string]*Device)
	}
	t.Devices[device.ID] = device
	t.UpdatedAt = time.Now()
}

// AddConnection adds a connection to the topology
func (t *NetworkTopology) AddConnection(conn *DeviceConnection) {
	if t.Connections == nil {
		t.Connections = make(map[string]*DeviceConnection)
	}
	t.Connections[conn.ID] = conn
	t.UpdatedAt = time.Now()
}

// GetDevice retrieves a device by ID
func (t *NetworkTopology) GetDevice(deviceID string) *Device {
	return t.Devices[deviceID]
}

// GetConnection retrieves a connection by ID
func (t *NetworkTopology) GetConnection(connID string) *DeviceConnection {
	return t.Connections[connID]
}

// RemoveDevice removes a device from the topology
func (t *NetworkTopology) RemoveDevice(deviceID string) {
	delete(t.Devices, deviceID)
	t.UpdatedAt = time.Now()
}

// RemoveConnection removes a connection from the topology
func (t *NetworkTopology) RemoveConnection(connID string) {
	delete(t.Connections, connID)
	t.UpdatedAt = time.Now()
}
