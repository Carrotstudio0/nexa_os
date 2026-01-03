package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

var SERVER_PORT = ":1413"

// Protocol Commands
const (
	CMD_PING    = "PING"
	CMD_FETCH   = "FETCH"
	CMD_PUBLISH = "PUBLISH"
	CMD_LIST    = "LIST"
)

// Response Codes
const (
	STATUS_OK           = 200
	STATUS_CREATED      = 201
	STATUS_BAD_REQ      = 400
	STATUS_NOT_FOUND    = 404
	STATUS_SERVER_ERROR = 500
)

// request for the protocol
type Request struct {
	Command string
	Target  string
	Body    string
}

// response for the protocol
type Response struct {
	Status  int
	Message string
	Body    string
}

// in-memory storage
var storage = make(map[string]string)

func main() {

	fmt.Println("Server running on port:", SERVER_PORT)

	ln, err := net.Listen("tcp", SERVER_PORT)
	if err != nil {
		panic(err)
	}

	defer ln.Close()

	fmt.Println("Server is ready, Waiting for connections")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept Error: ", err)
			// cause it's a while loop
			continue
		}

		fmt.Println("Net connection from: ", conn.RemoteAddr())
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	defer fmt.Printf("Connection closed: %s\n", conn.RemoteAddr())

	reader := bufio.NewReader(conn)

	for {

		// set request read timeout
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// read request line
		line, err := reader.ReadString('\n')
		if err != nil {
			return // connection closed or error
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fmt.Printf("Recived: %s\n", line)

		// parse request
		req := parseRequest(line)

		// handle request
		resp := handleRequest(req)

		// Send response
		sendReponse(conn, resp)

	}
}

// parse function for parsing protocol message
// e.g ->
// PING
// FETCH mypage
// PUBLISH mypage Hello World

// Format: COMMAND target [body]

func parseRequest(line string) Request {
	parts := strings.SplitN(line, " ", 3)

	req := Request{
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

// proccess requests and handle responses

func handleRequest(req Request) Response {
	switch req.Command {
	case CMD_PING:
		return Response{
			Status:  STATUS_OK,
			Message: "PONG!",
			Body:    fmt.Sprintf("Server time: %s", time.Now().Format(time.RFC3339)),
		}

	case CMD_FETCH:
		if req.Target == "" {
			return Response{
				Status:  STATUS_BAD_REQ,
				Message: "Target required",
			}
		}

		data, exists := storage[req.Target]
		if !exists {
			return Response{
				Status:  STATUS_NOT_FOUND,
				Message: "Not Found",
				Body:    fmt.Sprintf("Resource '%s' dose not exists", req.Target),
			}
		}

		return Response{
			Status:  STATUS_OK,
			Message: "Success",
			Body:    data,
		}

	case CMD_PUBLISH:
		if req.Target == "" {
			return Response{
				Status:  STATUS_BAD_REQ,
				Message: "Target required",
			}
		}

		storage[req.Target] = req.Body

		return Response{
			Status:  STATUS_CREATED,
			Message: "Published",
			Body:    fmt.Sprintf("Stored '%s' (%d bytes)", req.Target, len(req.Body)),
		}

	case CMD_LIST:
		var items []string
		for key := range storage {
			items = append(items, key)
		}

		if len(items) == 0 {
			return Response{
				Status:  STATUS_OK,
				Message: "Empty",
				Body:    "No items stored",
			}
		}

		return Response{
			Status:  STATUS_OK,
			Message: "Sucess",
			Body:    strings.Join(items, "\n"),
		}

	default:
		return Response{
			Status:  STATUS_BAD_REQ,
			Message: "uknown command",
			Body:    fmt.Sprintf("Command '%s' not recognized", req.Command),
		}

	}

}

func sendReponse(conn net.Conn, resp Response) {
	response := fmt.Sprintf("%d %s\n%s\n---END---\n", resp.Status, resp.Message, resp.Body)

	conn.Write([]byte(response))
	fmt.Printf("sent: %d %s\n", resp.Status, resp.Message)
}
