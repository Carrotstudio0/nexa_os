package network

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// ConnectionConfig holds configuration for device connections
type ConnectionConfig struct {
	ConnectionType    ConnectionType
	Timeout           time.Duration
	MaxRetries        int
	HeartbeatInterval time.Duration
	ReconnectWaitTime time.Duration
	TLSConfig         *tls.Config
}

// ConnectionHandler manages individual device connections
type ConnectionHandler struct {
	mu              sync.RWMutex
	device          *Device
	config          ConnectionConfig
	connection      net.Conn
	isConnected     bool
	lastMessageTime time.Time
	messageQueue    chan []byte
	errorChan       chan error
	stopChan        chan bool
	totalMsgs       int64
	errorMsgs       int64
}

// NewConnectionHandler creates a new connection handler for a device
func NewConnectionHandler(device *Device, config ConnectionConfig) *ConnectionHandler {
	return &ConnectionHandler{
		device:          device,
		config:          config,
		messageQueue:    make(chan []byte, 100),
		errorChan:       make(chan error, 10),
		stopChan:        make(chan bool),
		lastMessageTime: time.Now(),
	}
}

// Connect establishes a connection to the device
func (ch *ConnectionHandler) Connect() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if ch.isConnected {
		return fmt.Errorf("already connected to device %s", ch.device.ID)
	}

	var conn net.Conn
	var err error

	address := net.JoinHostPort(ch.device.IPAddress, fmt.Sprintf("%d", ch.device.Port))

	switch ch.config.ConnectionType {
	case ConnectionWiFi, ConnectionHotspot:
		// Standard TCP/TLS connection
		if ch.config.TLSConfig != nil {
			conn, err = tls.Dial("tcp", address, ch.config.TLSConfig)
		} else {
			conn, err = net.DialTimeout("tcp", address, ch.config.Timeout)
		}

	case ConnectionBluetooth:
		// Bluetooth connection (platform-specific implementation)
		// For now, we'll treat it as a TCP connection on a specific port
		btPort := ch.device.Port + 1000
		btAddress := net.JoinHostPort(ch.device.IPAddress, fmt.Sprintf("%d", btPort))
		conn, err = net.DialTimeout("tcp", btAddress, ch.config.Timeout)

	case ConnectionWiFiDirect:
		// WiFi Direct connection
		conn, err = net.DialTimeout("tcp", address, ch.config.Timeout)

	case ConnectionMesh:
		// Mesh connection through intermediate node
		conn, err = net.DialTimeout("tcp", address, ch.config.Timeout)

	default:
		return fmt.Errorf("unsupported connection type: %s", ch.config.ConnectionType)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to device %s: %v", ch.device.ID, err)
	}

	ch.connection = conn
	ch.isConnected = true
	ch.device.UpdateOnlineStatus(true)
	ch.lastMessageTime = time.Now()

	return nil
}

// Disconnect closes the connection to the device
func (ch *ConnectionHandler) Disconnect() error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	if !ch.isConnected || ch.connection == nil {
		return fmt.Errorf("not connected to device %s", ch.device.ID)
	}

	ch.stopChan <- true
	err := ch.connection.Close()
	ch.isConnected = false
	ch.device.UpdateOnlineStatus(false)

	return err
}

// SendMessage sends a message to the connected device
func (ch *ConnectionHandler) SendMessage(data interface{}) error {
	ch.mu.RLock()
	if !ch.isConnected {
		ch.mu.RUnlock()
		return fmt.Errorf("not connected to device %s", ch.device.ID)
	}
	conn := ch.connection
	ch.mu.RUnlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// Send message length first (4 bytes)
	lengthPrefix := make([]byte, 4)
	lengthPrefix[0] = byte(len(jsonData))
	lengthPrefix[1] = byte(len(jsonData) >> 8)
	lengthPrefix[2] = byte(len(jsonData) >> 16)
	lengthPrefix[3] = byte(len(jsonData) >> 24)

	if _, err := conn.Write(lengthPrefix); err != nil {
		ch.handleConnectionError(err)
		return fmt.Errorf("failed to send message length: %v", err)
	}

	if _, err := conn.Write(jsonData); err != nil {
		ch.handleConnectionError(err)
		return fmt.Errorf("failed to send message: %v", err)
	}

	ch.mu.Lock()
	ch.lastMessageTime = time.Now()
	ch.totalMsgs++
	ch.mu.Unlock()

	return nil
}

// ReceiveMessage receives a message from the connected device
func (ch *ConnectionHandler) ReceiveMessage(timeout time.Duration) ([]byte, error) {
	ch.mu.RLock()
	if !ch.isConnected {
		ch.mu.RUnlock()
		return nil, fmt.Errorf("not connected to device %s", ch.device.ID)
	}
	conn := ch.connection
	ch.mu.RUnlock()

	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(timeout))
	}

	// Read message length
	lengthPrefix := make([]byte, 4)
	if _, err := conn.Read(lengthPrefix); err != nil {
		ch.handleConnectionError(err)
		return nil, fmt.Errorf("failed to read message length: %v", err)
	}

	length := int(lengthPrefix[0]) |
		int(lengthPrefix[1])<<8 |
		int(lengthPrefix[2])<<16 |
		int(lengthPrefix[3])<<24

	if length <= 0 || length > 10*1024*1024 { // 10MB max
		return nil, fmt.Errorf("invalid message length: %d", length)
	}

	// Read message data
	data := make([]byte, length)
	bytesRead := 0
	for bytesRead < length {
		n, err := conn.Read(data[bytesRead:])
		if err != nil {
			ch.handleConnectionError(err)
			return nil, fmt.Errorf("failed to read message data: %v", err)
		}
		bytesRead += n
	}

	ch.mu.Lock()
	ch.lastMessageTime = time.Now()
	ch.totalMsgs++
	ch.mu.Unlock()

	return data, nil
}

// MeasureLatency measures the round-trip time to the device
func (ch *ConnectionHandler) MeasureLatency() (time.Duration, error) {
	start := time.Now()

	heartbeat := map[string]interface{}{
		"type":      "ping",
		"device_id": ch.device.ID,
		"timestamp": start.UnixNano(),
	}

	if err := ch.SendMessage(heartbeat); err != nil {
		return 0, err
	}

	// In a real system, we would wait for a "pong" response.
	// For this simulation/implementation, we'll assume the local processing is negligible
	// or that the successful write itself is enough to confirm connectivity.
	// Since we are using TCP, we can't easily get the RTT without a protocol-level ACK.
	latency := time.Since(start)

	ch.mu.Lock()
	ch.device.Metrics.LatencyMS = latency.Milliseconds()
	ch.mu.Unlock()

	return latency, nil
}

// StartHeartbeat starts sending periodic heartbeat messages
func (ch *ConnectionHandler) StartHeartbeat() {
	go func() {
		ticker := time.NewTicker(ch.config.HeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ch.stopChan:
				return
			case <-ticker.C:
				heartbeat := map[string]interface{}{
					"type":      "heartbeat",
					"device_id": ch.device.ID,
					"timestamp": time.Now().Unix(),
				}
				if err := ch.SendMessage(heartbeat); err != nil {
					ch.errorChan <- err
					ch.attemptReconnect()
				}
			}
		}
	}()
}

// IsConnected returns the connection status
func (ch *ConnectionHandler) IsConnected() bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.isConnected
}

// GetLastMessageTime returns the time of the last message
func (ch *ConnectionHandler) GetLastMessageTime() time.Time {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	return ch.lastMessageTime
}

// handleConnectionError handles connection errors
func (ch *ConnectionHandler) handleConnectionError(err error) {
	ch.mu.Lock()
	ch.isConnected = false
	ch.device.UpdateOnlineStatus(false)
	ch.errorMsgs++
	if ch.totalMsgs > 0 {
		ch.device.Metrics.ErrorRate = (float64(ch.errorMsgs) / float64(ch.totalMsgs)) * 100
	}
	ch.mu.Unlock()
	ch.errorChan <- err
}

// attemptReconnect attempts to reconnect to the device
func (ch *ConnectionHandler) attemptReconnect() {
	for i := 0; i < ch.config.MaxRetries; i++ {
		time.Sleep(ch.config.ReconnectWaitTime * time.Duration(i+1))

		if err := ch.Connect(); err == nil {
			return // Successfully reconnected
		}
	}

	ch.mu.Lock()
	ch.device.UpdateOnlineStatus(false)
	ch.mu.Unlock()
}

// Message represents a protocol message
type Message struct {
	Type      string                 `json:"type"`
	DeviceID  string                 `json:"device_id"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// HeartbeatMessage represents a heartbeat message
type HeartbeatMessage struct {
	Type            string `json:"type"`
	DeviceID        string `json:"device_id"`
	Timestamp       int64  `json:"timestamp"`
	Status          string `json:"status"`
	SignalStrength  int    `json:"signal_strength"`
	AvailableMemory int64  `json:"available_memory"`
}

// CommandMessage represents a command message
type CommandMessage struct {
	Type      string                 `json:"type"`
	DeviceID  string                 `json:"device_id"`
	Timestamp int64                  `json:"timestamp"`
	Command   string                 `json:"command"`
	Args      map[string]interface{} `json:"args,omitempty"`
}
