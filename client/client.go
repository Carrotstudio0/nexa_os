package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const DNS_SERVER = "localhost:1112"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	args := os.Args[1:]

	// Check if target is a .nexa name (needs DNS resolution)
	var target string
	var needsDNS bool

	if len(args) > 1 {
		target = args[1]
		needsDNS = strings.HasSuffix(target, ".nexa")
	}

	// Resolve DNS if needed
	var serverAddr string
	if needsDNS {
		fmt.Printf("Resolving %s via DNS...\n", target)
		addr, err := resolveDNS(target)
		if err != nil {
			fmt.Printf("DNS resolution failed: %v\n", err)
			return
		}
		serverAddr = addr
		fmt.Printf("Resolved to %s\n", serverAddr)
	} else {
		// Default server
		serverAddr = "localhost:1413"
	}

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Printf("Connection to %s failed: %v\n", serverAddr, err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to Nexa Server")

	command := strings.Join(args, " ")

	// send command
	fmt.Printf("Sending: %s\n", command)
	fmt.Fprintf(conn, "%s\n", command)

	// read response
	reader := bufio.NewReader(conn)

	// read status line
	statusLine, _ := reader.ReadString('\n')
	fmt.Printf("statusLine: %s", statusLine)

	// read body until ---END---
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if strings.TrimSpace(line) == "---END---" {
			break
		}

		fmt.Print(line)
	}
}

// resolveDNS queries the DNS server and returns IP:Port
func resolveDNS(name string) (string, error) {
	conn, err := net.Dial("tcp", DNS_SERVER)
	if err != nil {
		return "", fmt.Errorf("DNS server unreachable: %v", err)
	}
	defer conn.Close()

	// Send RESOLVE query
	query := fmt.Sprintf("RESOLVE %s\n", name)
	conn.Write([]byte(query))

	// Read response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	response = strings.TrimSpace(response)

	// Parse response: "200 RESOLVED 127.0.0.1:1413|service=web|ip=127.0.0.1"
	parts := strings.SplitN(response, " ", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid DNS response")
	}

	code := parts[0]
	if code != "200" {
		return "", fmt.Errorf("DNS error: %s", response)
	}

	// Extract IP:Port from body
	body := parts[2]
	addr := strings.Split(body, "|")[0]

	return addr, nil
}

func printUsage() {
	fmt.Println("Nexa Client - Usage:")
	fmt.Println()
	fmt.Println("Direct Commands (localhost:1413):")
	fmt.Println("  ./client PING")
	fmt.Println("  ./client FETCH <name>")
	fmt.Println("  ./client PUBLISH <name> <content>")
	fmt.Println("  ./client LIST")
	fmt.Println()
	fmt.Println("DNS-Resolved Commands (.nexa names):")
	fmt.Println("  ./client FETCH <name.nexa>")
	fmt.Println("  ./client PUBLISH <name.nexa> <content>")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ./client PING")
	fmt.Println("  ./client PUBLISH homepage Welcome to my site!")
	fmt.Println("  ./client FETCH homepage")
	fmt.Println("  ./client FETCH test.nexa")
	fmt.Println("  ./client PUBLISH test.nexa Hello World")
	fmt.Println("  ./client LIST")
	fmt.Println()
	fmt.Println("Note: Names ending with .nexa are automatically resolved via DNS")
}
