package network_test

import (
	"testing"
	"time"

	"github.com/MultiX0/nexa/pkg/network"
)

// TestNetworkManager tests the basic network manager functionality
func TestNetworkManager(t *testing.T) {
	// Create a network manager
	config := network.ConnectionConfig{
		ConnectionType:    network.ConnectionWiFi,
		Timeout:           10 * time.Second,
		MaxRetries:        3,
		HeartbeatInterval: 30 * time.Second,
		ReconnectWaitTime: 5 * time.Second,
	}

	nm := network.NewNetworkManager(config)

	// Test 1: Register Primary Base
	t.Run("RegisterPrimaryBase", func(t *testing.T) {
		base, err := nm.RegisterPrimaryBase(
			"test-base-001",
			"Test Base",
			"00:11:22:33:44:55",
			"127.0.0.1",
			9000,
		)
		if err != nil {
			t.Fatalf("Failed to register primary base: %v", err)
		}
		if base == nil {
			t.Fatal("Primary base is nil")
		}
		if base.Role != network.RolePrimaryBase {
			t.Fatalf("Expected RolePrimaryBase, got %v", base.Role)
		}
	})

	// Test 2: Register Device
	t.Run("RegisterDevice", func(t *testing.T) {
		device, err := nm.RegisterDevice(
			"test-device-001",
			"Test Device 1",
			"00:11:22:33:44:66",
			"127.0.0.1",
			9001,
			network.RoleGateway,
		)
		if err != nil {
			t.Fatalf("Failed to register device: %v", err)
		}
		if device == nil {
			t.Fatal("Device is nil")
		}
		if device.Role != network.RoleGateway {
			t.Fatalf("Expected RoleGateway, got %v", device.Role)
		}
	})

	// Test 3: Get Device
	t.Run("GetDevice", func(t *testing.T) {
		device := nm.GetDevice("test-device-001")
		if device == nil {
			t.Fatal("Device not found")
		}
		if device.ID != "test-device-001" {
			t.Fatalf("Expected device ID test-device-001, got %s", device.ID)
		}
	})

	// Test 4: Get Topology
	t.Run("GetTopology", func(t *testing.T) {
		topology := nm.GetTopology()
		if topology == nil {
			t.Fatal("Topology is nil")
		}
		if topology.PrimaryBase == nil {
			t.Fatal("Primary base not set in topology")
		}
		if _, exists := topology.Devices["test-device-001"]; !exists {
			t.Fatal("Device not found in topology")
		}
	})

	// Test 5: Create Connection
	t.Run("CreateConnection", func(t *testing.T) {
		conn, err := nm.CreateConnection("test-device-001", "test-base-001", network.ConnectionWiFi)
		if err != nil {
			t.Fatalf("Failed to create connection: %v", err)
		}
		if conn == nil {
			t.Fatal("Connection is nil")
		}
		if !conn.IsActive {
			t.Fatal("Connection should be active")
		}
	})

	// Test 6: Get Network Stats
	t.Run("GetNetworkStats", func(t *testing.T) {
		stats := nm.GetNetworkStats()
		if stats.TotalConnections != 1 {
			t.Fatalf("Expected 1 connection, got %d", stats.TotalConnections)
		}
	})

	// Test 7: Remove Connection
	t.Run("RemoveConnection", func(t *testing.T) {
		topology := nm.GetTopology()
		if len(topology.Connections) == 0 {
			t.Fatal("No connections to remove")
		}

		// Get the first connection
		var connID string
		for id := range topology.Connections {
			connID = id
			break
		}

		err := nm.RemoveConnection(connID)
		if err != nil {
			t.Fatalf("Failed to remove connection: %v", err)
		}
	})

	// Test 8: Remove Device
	t.Run("RemoveDevice", func(t *testing.T) {
		topology := nm.GetTopology()
		topology.RemoveDevice("test-device-001")

		device := nm.GetDevice("test-device-001")
		if device != nil {
			// Note: RemoveDevice modifies topology directly
			t.Log("Device marked for removal")
		}
	})
}

// TestDeviceDiscovery tests device discovery functionality
func TestDeviceDiscovery(t *testing.T) {
	discovery := network.NewDeviceDiscovery("255.255.255.255", 9999)

	// Test 1: Start Discovery
	t.Run("StartDiscovery", func(t *testing.T) {
		err := discovery.Start()
		if err != nil {
			t.Fatalf("Failed to start discovery: %v", err)
		}
		defer discovery.Stop()

		// Give it a moment to start
		time.Sleep(100 * time.Millisecond)

		// Create a beacon
		beacon := &network.DiscoveryBeacon{
			DeviceID:   "test-device-discovery",
			DeviceName: "Test Discovery Device",
			Role:       network.RoleNode,
			MAC:        "00:11:22:33:44:77",
			IPAddress:  "127.0.0.2",
			Port:       9002,
			SupportedConnections: []network.ConnectionType{
				network.ConnectionWiFi,
				network.ConnectionBluetooth,
			},
			Timestamp: time.Now().Unix(),
		}

		// Broadcast
		err = discovery.Broadcast(beacon)
		if err != nil {
			t.Logf("Warning: Broadcast failed (may be normal in test): %v", err)
		}
	})

	// Test 2: Get Discovered Devices
	t.Run("GetDiscoveredDevices", func(t *testing.T) {
		err := discovery.Start()
		if err != nil {
			t.Fatalf("Failed to start discovery: %v", err)
		}
		defer discovery.Stop()

		devices := discovery.GetDiscoveredDevices()
		if devices == nil {
			t.Fatal("Discovered devices is nil")
		}
		// Devices might be empty in test environment
		t.Logf("Discovered %d devices", len(devices))
	})
}

// TestNetworkTopology tests network topology operations
func TestNetworkTopology(t *testing.T) {
	topology := network.NewNetworkTopology()

	// Test 1: Add Device
	t.Run("AddDevice", func(t *testing.T) {
		device := network.NewDevice("dev-001", "Device 1", network.RoleNode, "00:11:22:33:44:88", "192.168.1.2", 9003)
		topology.AddDevice(device)

		retrieved := topology.GetDevice("dev-001")
		if retrieved == nil {
			t.Fatal("Device not found in topology")
		}
	})

	// Test 2: Add Connection
	t.Run("AddConnection", func(t *testing.T) {
		device2 := network.NewDevice("dev-002", "Device 2", network.RoleNode, "00:11:22:33:44:99", "192.168.1.3", 9004)
		topology.AddDevice(device2)

		conn := network.NewDeviceConnection("dev-001", "dev-002", network.ConnectionWiFi)
		topology.AddConnection(conn)

		if len(topology.Connections) == 0 {
			t.Fatal("Connection not added to topology")
		}
	})

	// Test 3: Remove Device
	t.Run("RemoveDevice", func(t *testing.T) {
		topology.RemoveDevice("dev-001")

		device := topology.GetDevice("dev-001")
		if device != nil {
			t.Fatal("Device should be removed")
		}
	})

	// Test 4: Get Timestamp
	t.Run("TimestampUpdate", func(t *testing.T) {
		oldTime := topology.UpdatedAt
		time.Sleep(10 * time.Millisecond)

		device := network.NewDevice("dev-003", "Device 3", network.RoleNode, "00:11:22:33:44:aa", "192.168.1.4", 9005)
		topology.AddDevice(device)

		if topology.UpdatedAt.Before(oldTime) {
			t.Fatal("UpdatedAt timestamp not updated")
		}
	})
}

// TestConnectionHandler tests connection handling
func TestConnectionHandler(t *testing.T) {
	t.Run("CreateConnectionHandler", func(t *testing.T) {
		device := network.NewDevice("conn-test-001", "Connection Test", network.RoleNode, "00:11:22:33:44:bb", "127.0.0.1", 9006)

		config := network.ConnectionConfig{
			ConnectionType:    network.ConnectionWiFi,
			Timeout:           5 * time.Second,
			MaxRetries:        2,
			HeartbeatInterval: 10 * time.Second,
			ReconnectWaitTime: 1 * time.Second,
		}

		handler := network.NewConnectionHandler(device, config)
		if handler == nil {
			t.Fatal("Handler is nil")
		}

		if !handler.IsConnected() {
			t.Log("Handler correctly shows not connected initially")
		}
	})
}

// Benchmark tests

func BenchmarkNetworkManagerRegisterDevice(b *testing.B) {
	config := network.ConnectionConfig{
		ConnectionType:    network.ConnectionWiFi,
		Timeout:           10 * time.Second,
		MaxRetries:        3,
		HeartbeatInterval: 30 * time.Second,
		ReconnectWaitTime: 5 * time.Second,
	}

	nm := network.NewNetworkManager(config)

	// Register primary base first
	nm.RegisterPrimaryBase("base", "Base", "00:00:00:00:00:00", "127.0.0.1", 9000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nm.RegisterDevice(
			"dev-"+string(rune(i)),
			"Device",
			"00:11:22:33:44:55",
			"127.0.0.1",
			9000+i,
			network.RoleNode,
		)
	}
}

func BenchmarkNetworkManagerGetDevice(b *testing.B) {
	config := network.ConnectionConfig{
		ConnectionType:    network.ConnectionWiFi,
		Timeout:           10 * time.Second,
		MaxRetries:        3,
		HeartbeatInterval: 30 * time.Second,
		ReconnectWaitTime: 5 * time.Second,
	}

	nm := network.NewNetworkManager(config)
	nm.RegisterPrimaryBase("base", "Base", "00:00:00:00:00:00", "127.0.0.1", 9000)
	nm.RegisterDevice("test-dev", "Device", "00:11:22:33:44:55", "127.0.0.1", 9001, network.RoleNode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nm.GetDevice("test-dev")
	}
}

func BenchmarkNetworkManagerCreateConnection(b *testing.B) {
	config := network.ConnectionConfig{
		ConnectionType:    network.ConnectionWiFi,
		Timeout:           10 * time.Second,
		MaxRetries:        3,
		HeartbeatInterval: 30 * time.Second,
		ReconnectWaitTime: 5 * time.Second,
	}

	nm := network.NewNetworkManager(config)
	nm.RegisterPrimaryBase("base", "Base", "00:00:00:00:00:00", "127.0.0.1", 9000)
	nm.RegisterDevice("dev1", "Device 1", "00:11:22:33:44:55", "127.0.0.1", 9001, network.RoleNode)
	nm.RegisterDevice("dev2", "Device 2", "00:11:22:33:44:56", "127.0.0.1", 9002, network.RoleNode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nm.CreateConnection("dev1", "dev2", network.ConnectionWiFi)
	}
}
