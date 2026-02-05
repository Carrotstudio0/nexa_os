package gateway

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/MultiX0/nexa/pkg/network"
)

// NetworkExpansionManager handles network expansion and device connectivity
type NetworkExpansionManager struct {
	mu              sync.RWMutex
	networkManager  *network.NetworkManager
	discovery       *network.DeviceDiscovery
	deviceRegistry  map[string]*network.Device
	relayRoutes     map[string]*RelayRoute
	stopChan        chan bool
	isRunning       bool
}

// RelayRoute represents a relay route for data forwarding
type RelayRoute struct {
	ID          string                `json:"id"`
	SourceID    string                `json:"source_id"`
	TargetID    string                `json:"target_id"`
	IntermediateID string              `json:"intermediate_id"`
	Priority    int                   `json:"priority"`
	Active      bool                  `json:"active"`
	CreatedAt   time.Time             `json:"created_at"`
}

// NewNetworkExpansionManager creates a new network expansion manager
func NewNetworkExpansionManager(networkMgr *network.NetworkManager, discoveryPort int) *NetworkExpansionManager {
	discovery := network.NewDeviceDiscovery("255.255.255.255", discoveryPort)

	return &NetworkExpansionManager{
		networkManager: networkMgr,
		discovery:      discovery,
		deviceRegistry: make(map[string]*network.Device),
		relayRoutes:    make(map[string]*RelayRoute),
		stopChan:       make(chan bool, 1),
		isRunning:      false,
	}
}

// Start starts the network expansion manager
func (nem *NetworkExpansionManager) Start() error {
	nem.mu.Lock()
	if nem.isRunning {
		nem.mu.Unlock()
		return fmt.Errorf("network expansion already running")
	}
	nem.isRunning = true
	nem.mu.Unlock()

	// Start device discovery
	if err := nem.discovery.Start(); err != nil {
		nem.mu.Lock()
		nem.isRunning = false
		nem.mu.Unlock()
		return fmt.Errorf("failed to start discovery: %v", err)
	}

	// Set discovery callback
	nem.discovery.SetOnDiscovered(func(resp *network.DiscoveryResponse) {
		nem.onDeviceDiscovered(resp)
	})

	// Start monitoring routine
	go nem.expansionLoop()

	return nil
}

// Stop stops the network expansion manager
func (nem *NetworkExpansionManager) Stop() {
	nem.mu.Lock()
	if !nem.isRunning {
		nem.mu.Unlock()
		return
	}
	nem.isRunning = false
	nem.mu.Unlock()

	nem.stopChan <- true
	nem.discovery.Stop()
}

// BroadcastDiscovery broadcasts discovery beacon
func (nem *NetworkExpansionManager) BroadcastDiscovery(device *network.Device) error {
	beacon := &network.DiscoveryBeacon{
		DeviceID:   device.ID,
		DeviceName: device.Name,
		Role:       device.Role,
		MAC:        device.MAC,
		IPAddress:  device.IPAddress,
		Port:       device.Port,
		SupportedConnections: []network.ConnectionType{
			network.ConnectionWiFi,
			network.ConnectionBluetooth,
			network.ConnectionWiFiDirect,
			network.ConnectionMesh,
		},
		Timestamp: time.Now().Unix(),
	}

	return nem.discovery.Broadcast(beacon)
}

// onDeviceDiscovered handles device discovery
func (nem *NetworkExpansionManager) onDeviceDiscovered(resp *network.DiscoveryResponse) {
	nem.mu.Lock()
	_, exists := nem.deviceRegistry[resp.DeviceID]
	nem.mu.Unlock()

	if !exists {
		log.Printf("[Network Expansion] Device discovered: %s (%s)", resp.DeviceName, resp.DeviceID)

		// Register the device
		device, err := nem.networkManager.RegisterDevice(
			resp.DeviceID,
			resp.DeviceName,
			resp.MAC,
			resp.IPAddress,
			resp.Port,
			resp.Role,
		)
		if err != nil {
			log.Printf("[Network Expansion] Failed to register device %s: %v", resp.DeviceID, err)
			return
		}

		nem.mu.Lock()
		nem.deviceRegistry[resp.DeviceID] = device
		nem.mu.Unlock()

		// Try to connect with best available method
		for _, connType := range resp.SupportedConnections {
			if err := nem.networkManager.ConnectDevice(resp.DeviceID, connType); err == nil {
				log.Printf("[Network Expansion] Connected to %s via %s", resp.DeviceID, connType)
				break
			}
		}
	}
}

// CreateRelayRoute creates a relay route between devices
func (nem *NetworkExpansionManager) CreateRelayRoute(sourceID, targetID, intermediateID string, priority int) (*RelayRoute, error) {
	nem.mu.Lock()
	defer nem.mu.Unlock()

	// Verify all devices exist
	source := nem.networkManager.GetDevice(sourceID)
	target := nem.networkManager.GetDevice(targetID)
	intermediate := nem.networkManager.GetDevice(intermediateID)

	if source == nil || target == nil || intermediate == nil {
		return nil, fmt.Errorf("one or more devices not found")
	}

	// Create relay route
	route := &RelayRoute{
		ID:             generateID(),
		SourceID:       sourceID,
		TargetID:       targetID,
		IntermediateID: intermediateID,
		Priority:       priority,
		Active:         true,
		CreatedAt:      time.Now(),
	}

	nem.relayRoutes[route.ID] = route

	// Create connections for the relay
	_, err1 := nem.networkManager.CreateConnection(sourceID, intermediateID, network.ConnectionMesh)
	_, err2 := nem.networkManager.CreateConnection(intermediateID, targetID, network.ConnectionMesh)

	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("failed to create relay connections")
	}

	log.Printf("[Network Expansion] Relay route created: %s -> %s -> %s", sourceID, intermediateID, targetID)

	return route, nil
}

// RelayMessage relays a message through an intermediate device
func (nem *NetworkExpansionManager) RelayMessage(sourceID, targetID, intermediateID string, data interface{}) error {
	nem.mu.RLock()
	// Find the relay route
	var route *RelayRoute
	for _, r := range nem.relayRoutes {
		if r.SourceID == sourceID && r.TargetID == targetID && r.IntermediateID == intermediateID {
			route = r
			break
		}
	}
	nem.mu.RUnlock()

	if route == nil || !route.Active {
		return fmt.Errorf("relay route not found or inactive")
	}

	// Create relay message
	relayMsg := map[string]interface{}{
		"type":             "relay",
		"source_id":        sourceID,
		"target_id":        targetID,
		"intermediate_id":  intermediateID,
		"data":             data,
		"timestamp":        time.Now().Unix(),
	}

	// Send to intermediate device
	return nem.networkManager.SendCommandToDevice(intermediateID, "relay_message", relayMsg)
}

// GetNetworkTopology returns the current network topology
func (nem *NetworkExpansionManager) GetNetworkTopology() *network.NetworkTopology {
	return nem.networkManager.GetTopology()
}

// GetNetworkStats returns network statistics
func (nem *NetworkExpansionManager) GetNetworkStats() network.NetworkStats {
	return nem.networkManager.GetNetworkStats()
}

// GetRelayRoutes returns all relay routes
func (nem *NetworkExpansionManager) GetRelayRoutes() map[string]*RelayRoute {
	nem.mu.RLock()
	defer nem.mu.RUnlock()

	routes := make(map[string]*RelayRoute)
	for id, route := range nem.relayRoutes {
		routes[id] = route
	}

	return routes
}

// RemoveRelayRoute removes a relay route
func (nem *NetworkExpansionManager) RemoveRelayRoute(routeID string) error {
	nem.mu.Lock()
	route, exists := nem.relayRoutes[routeID]
	if !exists {
		nem.mu.Unlock()
		return fmt.Errorf("relay route not found")
	}

	delete(nem.relayRoutes, routeID)
	nem.mu.Unlock()

	// Remove the connections
	topology := nem.networkManager.GetTopology()
	for connID, conn := range topology.Connections {
		if (conn.SourceDeviceID == route.SourceID && conn.TargetDeviceID == route.IntermediateID) ||
			(conn.SourceDeviceID == route.IntermediateID && conn.TargetDeviceID == route.TargetID) {
			nem.networkManager.RemoveConnection(connID)
		}
	}

	log.Printf("[Network Expansion] Relay route removed: %s", routeID)
	return nil
}

// expansionLoop is the main loop for network expansion
func (nem *NetworkExpansionManager) expansionLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-nem.stopChan:
			return
		case <-ticker.C:
			nem.maintainNetworkExpansion()
		}
	}
}

// maintainNetworkExpansion maintains the expanded network
func (nem *NetworkExpansionManager) maintainNetworkExpansion() {
	topology := nem.networkManager.GetTopology()

	// Check for isolated devices and create relay routes if needed
	for deviceID, device := range topology.Devices {
		if device.IsOnline {
			// Count active connections for this device
			activeConnCount := 0
			for _, conn := range topology.Connections {
				if (conn.SourceDeviceID == deviceID || conn.TargetDeviceID == deviceID) && conn.IsActive {
					activeConnCount++
				}
			}

			// If device has no active connections and is not primary base, try to create a relay
			if activeConnCount == 0 && device.Role != network.RolePrimaryBase && topology.PrimaryBase != nil {
				nem.findAndCreateRelayPath(deviceID, topology)
			}
		}
	}
}

// findAndCreateRelayPath finds and creates a relay path for isolated devices
func (nem *NetworkExpansionManager) findAndCreateRelayPath(deviceID string, topology *network.NetworkTopology) {
	// Find an intermediate gateway that is connected
	var bestGateway *network.Device

	for _, device := range topology.Devices {
		if device.Role == network.RoleGateway && device.IsOnline {
			// Count this gateway's active connections
			connCount := 0
			for _, conn := range topology.Connections {
				if (conn.SourceDeviceID == device.ID || conn.TargetDeviceID == device.ID) && conn.IsActive {
					connCount++
				}
			}

			if connCount > 0 {
				bestGateway = device
				break
			}
		}
	}

	if bestGateway != nil && topology.PrimaryBase != nil {
		_, err := nem.CreateRelayRoute(deviceID, topology.PrimaryBase.ID, bestGateway.ID, 1)
		if err == nil {
			log.Printf("[Network Expansion] Auto relay created for isolated device %s", deviceID)
		}
	}
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}

// GetNetworkExpansionAPI returns API endpoints for network expansion
func (nem *NetworkExpansionManager) GetNetworkExpansionAPI() map[string]interface{} {
	return map[string]interface{}{
		"description": "Network Expansion API v3.1",
		"endpoints": []map[string]string{
			{
				"method":      "GET",
				"path":        "/api/network/topology",
				"description": "Get current network topology",
			},
			{
				"method":      "GET",
				"path":        "/api/network/stats",
				"description": "Get network statistics",
			},
			{
				"method":      "GET",
				"path":        "/api/network/devices",
				"description": "Get all devices",
			},
			{
				"method":      "POST",
				"path":        "/api/network/relay",
				"description": "Create relay route",
			},
			{
				"method":      "GET",
				"path":        "/api/network/relay",
				"description": "Get all relay routes",
			},
			{
				"method":      "DELETE",
				"path":        "/api/network/relay/:routeId",
				"description": "Delete relay route",
			},
		},
	}
}

// SerializeNetworkState serializes the network state to JSON
func (nem *NetworkExpansionManager) SerializeNetworkState() ([]byte, error) {
	nem.mu.RLock()
	defer nem.mu.RUnlock()

	state := map[string]interface{}{
		"topology":     nem.networkManager.GetTopology(),
		"relay_routes": nem.relayRoutes,
		"stats":        nem.networkManager.GetNetworkStats(),
		"timestamp":    time.Now(),
	}

	return json.MarshalIndent(state, "", "  ")
}
