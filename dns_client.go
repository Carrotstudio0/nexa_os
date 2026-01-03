package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// Connect to DNS server
	conn, err := net.Dial("tcp", "localhost:1112")
	if err != nil {
		fmt.Printf("DNS connection failed: %v\n", err)
		fmt.Println("Make sure DNS server is running on port 1112")
		return
	}
	defer conn.Close()

	// Build query from arguments
	query := strings.Join(os.Args[1:], " ")

	// Send query
	fmt.Printf("DNS Query: %s\n", query)
	fmt.Fprintf(conn, "%s\n", query)

	// Read response
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Read error: %v\n", err)
		return
	}

	// Parse and display response
	response = strings.TrimSpace(response)
	parseResponse(response)
}

func parseResponse(response string) {
	parts := strings.SplitN(response, " ", 3)
	if len(parts) < 2 {
		fmt.Println(response)
		return
	}

	code := parts[0]
	status := parts[1]
	body := ""
	if len(parts) > 2 {
		body = parts[2]
	}

	// Color output based on status code
	switch code {
	case "200", "201":
		fmt.Printf("%s %s\n", code, status)
	case "404":
		fmt.Printf("%s %s\n", code, status)
	case "400", "409", "500":
		fmt.Printf("%s %s\n", code, status)
	default:
		fmt.Printf("%s %s\n", code, status)
	}

	if body != "" {
		// Special formatting for LIST command
		if strings.Contains(body, "|") && strings.Contains(body, "=") {
			fmt.Println("\n Records:")
			records := strings.Split(body, "|")
			for _, rec := range records {
				fmt.Printf("   • %s\n", rec)
			}
		} else {
			fmt.Printf("%s\n", body)
		}
	}
}

func printUsage() {
	fmt.Println("NEXA DNS Client - Usage:")
	fmt.Println()
	fmt.Println(" Query Commands:")
	fmt.Println("  ./dns_client RESOLVE <n>")
	fmt.Println("  ./dns_client LIST")
	fmt.Println("  ./dns_client PING")
	fmt.Println()
	fmt.Println(" Management Commands:")
	fmt.Println("  ./dns_client REGISTER <n> <ip> <port> <service>")
	fmt.Println("  ./dns_client UPDATE <n> <ip> <port> <service>")
	fmt.Println("  ./dns_client DELETE <n>")
	fmt.Println()
	fmt.Println(" Examples:")
	fmt.Println("  ./dns_client RESOLVE mysite.nexa")
	fmt.Println("  ./dns_client REGISTER myapp.nexa 192.168.1.100 8080 web")
	fmt.Println("  ./dns_client UPDATE myapp.nexa 192.168.1.101 8080 web")
	fmt.Println("  ./dns_client DELETE myapp.nexa")
	fmt.Println("  ./dns_client LIST")
}
