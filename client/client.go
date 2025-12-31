package client

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

	conn, err := net.Dial("tcp", "localhost:1413")
	if err != nil {
		fmt.Println("Connection failed: ", err.Error())
		return
	}

	defer conn.Close()

	fmt.Println("Connected to NetWorking Server")

	command := strings.Join(os.Args[1:], " ")

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

func printUsage() {
	fmt.Println("Networking Client - Usage:")
	fmt.Println("	./client PING")
	fmt.Println("	./client FETCH <name>")
	fmt.Println("	./client PUBLISH <name> <content>")
	fmt.Println("	./client LIST")
	fmt.Println()
	fmt.Println("Eamples:")
	fmt.Println("	./client PING")
	fmt.Println("	./client PUBLISH homepage Welcome to my site!")
	fmt.Println("	./client FETCH homepage")
	fmt.Println("	./client LIST")
}
