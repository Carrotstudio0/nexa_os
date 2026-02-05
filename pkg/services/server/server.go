package server

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/MultiX0/nexa/pkg/auth"
	"github.com/MultiX0/nexa/pkg/config"
	"github.com/MultiX0/nexa/pkg/governance"
	"github.com/MultiX0/nexa/pkg/ledger"
	"github.com/MultiX0/nexa/pkg/network"
	"github.com/MultiX0/nexa/pkg/nexa"
	"github.com/MultiX0/nexa/pkg/utils"
)

var (
	chain          *ledger.Blockchain
	authManager    *auth.AuthManager
	networkManager *network.NetworkManager
	govManager     *governance.GovernanceManager
)

func Start(nm *network.NetworkManager, gm *governance.GovernanceManager) {
	var err error
	networkManager = nm
	govManager = gm

	// Initialize Blockchain Ledger
	chain, err = ledger.NewBlockchain("ledger.json")
	if err != nil {
		utils.LogFatal("Server", "Failed to init ledger: "+err.Error())
	}

	// Initialize Auth
	usersFile := utils.FindFile("users.json")
	authManager, err = auth.NewAuthManager(usersFile)
	if err != nil {
		utils.LogFatal("Server", "Failed to init auth: "+err.Error())
	}

	// Metrics reporter
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		for range ticker.C {
			if networkManager != nil {
				networkManager.UpdateServiceMetrics("core", map[string]interface{}{
					"blocks":         len(chain.Chain),
					"ledger_size":    len(chain.Data),
					"last_heartbeat": time.Now().Format("15:04:05"),
				})
			}
		}
	}()

	localIP := utils.GetLocalIP()

	// TLS Configuration
	certFile, keyFile := utils.FindCertFiles()
	var serverTLSConfig *tls.Config
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Printf("Warning: Failed to load TLS certs: %v", err)
			serverTLSConfig = &tls.Config{}
		} else {
			serverTLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		}
	} else {
		serverTLSConfig = &tls.Config{}
	}

	// Start Listener
	ln, err := tls.Listen("tcp", "0.0.0.0:"+config.ServerPort, serverTLSConfig)
	if err != nil {
		utils.LogFatal("Server", "Listener failed: "+err.Error())
	}

	utils.LogInfo("Server", fmt.Sprintf("Blockchain Height: %d blocks", len(chain.Chain)))
	utils.LogInfo("Server", fmt.Sprintf("Listening Port:    %s", config.ServerPort))
	utils.SaveEndpoint("core", fmt.Sprintf("tcp://%s:%s", localIP, config.ServerPort))

	for {
		conn, err := ln.Accept()
		if err != nil {
			utils.LogWarning("Server", "Accept error: "+err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		conn.SetReadDeadline(time.Now().Add(120 * time.Second))
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		utils.LogInfo("Server", fmt.Sprintf("[%s] REQ: %s", conn.RemoteAddr(), line))

		req := parseRequest(line)
		resp := processRequest(req)

		sendResponse(conn, resp)
	}
}

func parseRequest(line string) nexa.Request {
	parts := strings.SplitN(line, " ", 3)
	req := nexa.Request{
		Command: strings.ToUpper(parts[0]),
	}
	if len(parts) > 1 {
		req.Target = parts[1]
	}
	if len(parts) > 2 {
		req.Body = parts[2]
	}
	return req
}

func processRequest(req nexa.Request) nexa.Response {
	switch req.Command {
	case nexa.CMD_PING:
		return nexa.Response{Status: nexa.STATUS_OK, Message: "PONG", Body: fmt.Sprintf("Height=%d", len(chain.Chain))}

	case nexa.CMD_FETCH:
		if req.Target == "" {
			return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: "Target Required"}
		}
		val, exists := chain.Get(req.Target)
		if !exists {
			return nexa.Response{Status: nexa.STATUS_NOT_FOUND, Message: "Not Found"}
		}
		return nexa.Response{Status: nexa.STATUS_OK, Message: "Success", Body: val}

	case nexa.CMD_PUBLISH:
		if req.Target == "" {
			return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: "Target Required"}
		}
		// In a real network, we'd verify signature. For now, we assume simple publishing.
		block := chain.AddBlock(req.Target, req.Body, "Node-Local")
		return nexa.Response{Status: nexa.STATUS_CREATED, Message: "Mined", Body: block.Hash}

	case nexa.CMD_LIST:
		// Return list of keys in data map
		keys := make([]string, 0, len(chain.Data))
		for k := range chain.Data {
			keys = append(keys, k)
		}
		return nexa.Response{Status: nexa.STATUS_OK, Message: "Keys", Body: strings.Join(keys, ",")}

	case "LEDGER": // New command to see chain info
		return nexa.Response{Status: nexa.STATUS_OK, Message: "Chain Info", Body: fmt.Sprintf("Height: %d, Valid: %v", len(chain.Chain), chain.IsChainValid())}

	case nexa.CMD_AUTH:
		parts := strings.Fields(req.Body)
		if len(parts) != 2 {
			return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: "Usage: AUTH user pass"}
		}
		valid, role := authManager.Verify(parts[0], parts[1])
		if !valid {
			if govManager != nil {
				govManager.ReportEvent("Security", governance.LevelWarning,
					"Failed Authentication Attempt",
					fmt.Sprintf("User: %s IP: Node-Remote", parts[0]),
					"Log and Monitor")
			}
			return nexa.Response{Status: nexa.STATUS_UNAUTHORIZED, Message: "Invalid Credentials"}
		}
		return nexa.Response{Status: nexa.STATUS_OK, Message: "Authenticated", Body: role}

	// Network Expansion Commands (v3.1)
	case "NETWORK":
		if req.Target == "TOPOLOGY" {
			topology := networkManager.GetTopology()
			data, _ := json.MarshalIndent(topology, "", "  ")
			return nexa.Response{Status: nexa.STATUS_OK, Message: "Network Topology", Body: string(data)}
		}
		if req.Target == "STATS" {
			stats := networkManager.GetNetworkStats()
			data, _ := json.MarshalIndent(stats, "", "  ")
			return nexa.Response{Status: nexa.STATUS_OK, Message: "Network Stats", Body: string(data)}
		}
		if req.Target == "DEVICES" {
			topology := networkManager.GetTopology()
			data, _ := json.MarshalIndent(topology.Devices, "", "  ")
			return nexa.Response{Status: nexa.STATUS_OK, Message: "Devices", Body: string(data)}
		}
		return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: "Unknown NETWORK subcommand"}

	case "CONNECT":
		// Format: CONNECT <deviceId> <connectionType>
		parts := strings.SplitN(req.Body, " ", 2)
		if len(parts) < 1 || parts[0] == "" {
			return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: "Usage: CONNECT <deviceId> [connectionType]"}
		}
		connType := network.ConnectionWiFi
		if len(parts) > 1 {
			connType = network.ConnectionType(parts[1])
		}
		if err := networkManager.ConnectDevice(parts[0], connType); err != nil {
			return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: fmt.Sprintf("Failed to connect: %v", err)}
		}
		return nexa.Response{Status: nexa.STATUS_OK, Message: "Device Connected", Body: parts[0]}

	case "DISCONNECT":
		if req.Body == "" {
			return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: "DISCONNECT <deviceId> required"}
		}
		if err := networkManager.DisconnectDevice(req.Body); err != nil {
			return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: fmt.Sprintf("Failed to disconnect: %v", err)}
		}
		return nexa.Response{Status: nexa.STATUS_OK, Message: "Device Disconnected", Body: req.Body}

	default:
		return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: "Unknown Command"}
	}
}

func sendResponse(conn net.Conn, resp nexa.Response) {
	fmt.Fprintf(conn, "%d %s\n%s\n---END---\n", resp.Status, resp.Message, resp.Body)
}
