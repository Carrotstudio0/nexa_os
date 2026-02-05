package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"os"
	"strings"
)

const (
	DNS_SERVER  = "localhost:1112"
	NEXA_SERVER = "localhost:1413"
)

func main() {
	if len(os.Args) < 2 {
		startInteractiveShell()
		return
	}
	handleCommand(os.Args[1:])
}

func startInteractiveShell() {
	fmt.Println("Nexa CLI Client v2.0 (Blockchain Enabled)")
	fmt.Println("Commands: PING, LIST, FETCH <key>, PUBLISH <key> <val>, LEDGER, AUTH, HELP, EXIT")
	fmt.Println("----------------------------------------")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("nexa> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "exit" || line == "quit" {
			break
		}
		if line == "" {
			continue
		}
		if line == "help" {
			printUsage()
			continue
		}
		handleCommand(strings.Fields(line))
	}
}

func handleCommand(args []string) {
	if len(args) == 0 {
		return
	}

	target := ""
	needsDNS := false

	if len(args) > 1 {
		target = args[1]
		needsDNS = strings.HasSuffix(target, ".nexa")
	}

	serverAddr := NEXA_SERVER

	if needsDNS {
		fmt.Printf("[DNS] Resolving %s...\n", target)
		addr, err := resolveDNS(target)
		if err != nil {
			fmt.Printf("[DNS] Error: %v\n", err)
			return
		}
		serverAddr = addr
		fmt.Printf("[DNS] Resolved to %s\n", serverAddr)
	}

	conn, err := tls.Dial("tcp", serverAddr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		fmt.Printf("[NET] Connection failed: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Fprintf(conn, "%s\n", strings.Join(args, " "))

	reader := bufio.NewReader(conn)
	status, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("[NET] Error reading response")
		return
	}
	fmt.Print(status)

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

func resolveDNS(name string) (string, error) {
	conn, err := tls.Dial("tcp", DNS_SERVER, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return "", fmt.Errorf("DNS unreachable")
	}
	defer conn.Close()

	fmt.Fprintf(conn, "RESOLVE %s\n", name)
	reader := bufio.NewReader(conn)
	resp, _ := reader.ReadString('\n')

	parts := strings.SplitN(resp, " ", 3)
	if len(parts) < 3 || parts[0] != "200" {
		return "", fmt.Errorf("lookup failed: %s", resp)
	}

	return strings.Split(parts[2], "|")[0], nil
}

func printUsage() {
	fmt.Println("  PING                     - Check server status")
	fmt.Println("  LIST                     - List all keys")
	fmt.Println("  FETCH <key>              - Get content from chain")
	fmt.Println("  PUBLISH <key> <content>  - Mine a new block with content")
	fmt.Println("  LEDGER                   - Show blockchain info")
	fmt.Println("  AUTH <user> <pass>       - Login to session")
}
