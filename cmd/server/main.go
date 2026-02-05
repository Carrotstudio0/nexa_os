package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/MultiX0/nexa/pkg/auth"
	"github.com/MultiX0/nexa/pkg/ledger"
	"github.com/MultiX0/nexa/pkg/nexa"
	"github.com/MultiX0/nexa/pkg/utils"
)

var (
	chain       *ledger.Blockchain
	authManager *auth.AuthManager
)

func main() {
	var err error

	// Initialize Blockchain Ledger
	chain, err = ledger.NewBlockchain("ledger.json")
	if err != nil {
		log.Fatalf("Failed to init ledger: %v", err)
	}

	// Initialize Auth
	usersFile := utils.FindFile("users.json")
	authManager, err = auth.NewAuthManager(usersFile)
	if err != nil {
		log.Fatalf("Failed to init auth: %v", err)
	}

	// TLS Configuration
	certFile, keyFile := utils.FindCertFiles()
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load TLS certs: %v", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}

	// Start Listener
	ln, err := tls.Listen("tcp", "0.0.0.0:"+nexa.PORT_SERVER, tlsConfig)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	// Professional Startup Banner
	fmt.Println(`
    _   _                      ____                                   
   | \ | | _____  ____ _      / ___|  ___ _ ____   _____ _ __ 
   |  \| |/ _ \ \/ / _' |     \___ \ / _ \ '__\ \ / / _ \ '__|
   | |\  |  __/>  < (_| |      ___) |  __/ |   \ V /  __/ |   
   |_| \_|\___/_/\_\__,_|     |____/ \___|_|    \_/ \___|_|   
                                                               v3.0 Ultimate`)
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Printf("   [INFO]  Initializing Core System...\n")
	fmt.Printf("   [INFO]  TLS Security:     %s\n", "ENABLED 🔒")
	fmt.Printf("   [INFO]  Blockchain Height: %d blocks\n", len(chain.Chain))
	fmt.Printf("   [INFO]  Listening Port:    %s\n", nexa.PORT_SERVER)
	fmt.Println("   ════════════════════════════════════════════════════════════════")
	fmt.Println("   ✅  SYSTEM READY FOR CONNECTIONS")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("   [TX-FAIL] Accept Error: %v", err)
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

		log.Printf("[%s] REQ: %s", conn.RemoteAddr(), line)

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
			return nexa.Response{Status: nexa.STATUS_UNAUTHORIZED, Message: "Invalid Credentials"}
		}
		return nexa.Response{Status: nexa.STATUS_OK, Message: "Authenticated", Body: role}

	default:
		return nexa.Response{Status: nexa.STATUS_BAD_REQ, Message: "Unknown Command"}
	}
}

func sendResponse(conn net.Conn, resp nexa.Response) {
	fmt.Fprintf(conn, "%d %s\n%s\n---END---\n", resp.Status, resp.Message, resp.Body)
}
