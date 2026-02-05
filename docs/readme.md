# Nexa Protocol

A custom network protocol built from scratch with its own DNS system. Think of it as building your own mini-internet that runs on top of TCP/IP.

## 🚀 Quick Start

**For the fastest setup:**
1. Run `test.bat` to verify your environment
2. Run `build.bat` to build all components
3. Run `start.bat` to launch all services
4. Open http://localhost:8080 (user: admin, pass: admin123)

📚 **Documentation:**
- [QUICKSTART.md](QUICKSTART.md) - Quick reference guide
- [BUILD.md](BUILD.md) - Detailed build instructions
- [readme.md](readme.md) - Full documentation (this file)

## What is this?

Nexa is a complete network stack that includes:

- **Custom Protocol**: Your own communication protocol (like HTTP, but yours)
- **DNS System**: Translate names like `mysite.nexa` into IP addresses and ports
- **Client/Server**: Tools to interact with the network

This isn't just a toy project - it's a fully functional alternative network layer that demonstrates how protocols and DNS actually work under the hood.

## Project Structure

```
nexa/
├── server/
│   └── server.go          # Main Nexa server (port 1413)
├── dns/
│   └── dns_server.go      # DNS resolution service (port 1112)
├── client/
│   └── client.go          # Smart client with DNS support
├── dns_client.go          # DNS management tool
└── go.mod
```


## TLS/SSL Encryption Support

**All connections (server, DNS, client) now use TLS encryption by default.**

- شهادات TLS ذاتية التوقيع موجودة في مجلد `certs/`
- عند التشغيل، جميع الأطراف تستخدم الاتصال المشفر (InsecureSkipVerify=true للعمل التجريبي)

## How it Works

### The Flow

1. **DNS Server** runs on port 1112 and maintains a registry of `.nexa` domains
2. **Nexa Server** runs on port 1413 and stores/serves content
3. **Client** can either:
   - Connect directly to the server using `localhost:1413`
   - Use `.nexa` domain names which get resolved via DNS first

### Example Flow

```
User runs: ./client FETCH mysite.nexa

1. Client sees ".nexa" extension
2. Client queries DNS server: "What's the address for mysite.nexa?"
3. DNS responds: "127.0.0.1:1413"
4. Client connects to that address
5. Client sends: FETCH mysite.nexa
6. Server returns the stored content
```

## Installation

### Prerequisites

- Go 1.16 or higher
- A terminal

### Build

```bash
# Clone or download the project
cd nexa

# Build all components
go build -o bin/server ./server/server.go
go build -o bin/dns ./dns/dns_server.go
go build -o bin/client ./client/client.go
go build -o bin/dns-client ./dns_client.go
```

Or just run them directly with `go run`.

## Quick Start


### Step 1: Start the DNS Server (TLS)

افتح نافذة طرفية:

```bash
cd dns
go run dns_server.go
```

يجب أن ترى:
```
DNS Server starting with TLS on :1112
--- DNS Server ready (TLS) ---
... (default records)
```

### Step 2: Start the Nexa Server (TLS)

افتح نافذة طرفية أخرى:

```bash
cd server
go run server.go
```

يجب أن ترى:
```
Server running with TLS on port: :1413
Server is ready with TLS, Waiting for connections
```

### Step 3: Use the Client (TLS)

افتح نافذة طرفية ثالثة وجرب الأوامر:

```bash
# اختبار الاتصال
go run client/client.go PING

# تخزين بيانات
go run client/client.go PUBLISH homepage "Welcome to Nexa"

# جلب البيانات
go run client/client.go FETCH homepage

# عرض كل البيانات
go run client/client.go LIST

# جلب بيانات عبر DNS
go run client/client.go FETCH mysite.nexa

# تخزين بيانات باسم DNS
go run client/client.go PUBLISH mysite.nexa "Hello from DNS"
```

## Protocol Commands

### Nexa Server Commands

The server understands these commands:

- **PING** - Health check, server responds with timestamp
- **FETCH <name>** - Retrieve stored content
- **PUBLISH <name> <content>** - Store content with a name
- **LIST** - Show all stored items

### DNS Commands

Manage the DNS registry:

- **PING** - Check if DNS server is alive
- **RESOLVE <name.nexa>** - Get IP:Port for a domain
- **REGISTER <name.nexa> <ip> <port> <service>** - Add new domain
- **UPDATE <name.nexa> <ip> <port> <service>** - Update existing domain
- **DELETE <name.nexa>** - Remove a domain
- **LIST** - Show all registered domains

## Usage Examples

### Working with Content

```bash
# Store a homepage
go run client/client.go PUBLISH homepage "Welcome to my site"

# Store multiple pages
go run client/client.go PUBLISH about "About page content"
go run client/client.go PUBLISH contact "email@example.com"

# Retrieve them
go run client/client.go FETCH homepage
go run client/client.go FETCH about

# See everything
go run client/client.go LIST
```

### Working with DNS

```bash
# See what domains exist
go run dns_client.go LIST

# Look up a domain
go run dns_client.go RESOLVE mysite.nexa

# Register a new domain
go run dns_client.go REGISTER blog.nexa 127.0.0.1 1413 web

# Now you can use it
go run client/client.go PUBLISH blog.nexa "My first post"
go run client/client.go FETCH blog.nexa

# Update a domain (maybe it moved to a different port)
go run dns_client.go UPDATE blog.nexa 127.0.0.1 1414 web

# Remove a domain
go run dns_client.go DELETE blog.nexa
```

### The Magic: DNS Resolution

When you use a `.nexa` domain, the client automatically:

1. Contacts the DNS server
2. Gets the real IP and port
3. Connects to that server
4. Sends your command

```bash
# This command triggers DNS resolution
go run client/client.go FETCH mysite.nexa

# Output shows:
# Resolving mysite.nexa via DNS...
# Resolved to 127.0.0.1:1413
# Connected to Nexa Server
# [content appears here]
```

## Understanding the Code

### Server (server/server.go)

The server is straightforward:
- Listens on TCP port 1413
- Accepts connections
- Parses incoming commands
- Stores data in memory (a simple map)
- Sends responses back

### DNS Server (dns/dns_server.go)

The DNS server:
- Listens on TCP port 1112
- Maintains a registry (map) of name -> address mappings
- Handles RESOLVE queries
- Allows registration/updates/deletion of domains
- Thread-safe with mutex locks

### Client (client/client.go)

The smart client:
- Takes command line arguments
- Checks if the target is a `.nexa` domain
- If yes: queries DNS first, then connects
- If no: connects directly to localhost:1413
- Sends the command and displays the response

### DNS Client (dns_client.go)

Simple DNS management tool:
- Sends DNS commands to the DNS server
- Formats and displays responses
- Used for managing the DNS registry

## Why This Matters

This project demonstrates:

1. **Protocol Design**: How protocols like HTTP actually work
2. **DNS Resolution**: How domain names get translated to addresses
3. **Client-Server Architecture**: The foundation of the internet
4. **Network Programming**: Working with TCP sockets in Go

You're not using any framework or library for the protocol itself - it's all raw TCP connections and string parsing. This is how the real internet works at a lower level.

## Limitations

- **In-Memory Storage**: Server data is lost on restart
- **No Encryption**: Everything is plain text
- **No Authentication**: Anyone can publish/fetch
- **Single-Threaded DNS**: One query at a time (though the server handles multiple connections)
- **Local Only**: Designed for localhost, but can work on LAN

## Future Ideas

- Add persistent storage (save to disk)
- Implement authentication
- Add TLS/encryption
- Support binary data transfer
- Build a simple web interface
- Make DNS distributed
- Add caching layers
- Support multiple server instances

## Troubleshooting

**"Connection refused" error:**
- Make sure the server is running first
- Check the port numbers match (1413 for server, 1112 for DNS)

**DNS resolution fails:**
- Ensure DNS server is running
- Verify the domain is registered with `go run dns_client.go LIST`

**"Not found" errors:**
- Use LIST command to see what's actually stored
- Remember: storage is in-memory, restart = data loss

## Technical Details

- **Language**: Go
- **Network**: TCP/IP
- **Protocol**: Custom text-based
- **DNS Port**: 1112
- **Server Port**: 1413
- **Response Format**: Status line + body + END marker

## License

This is a learning project. Use it however you want.

## Contributing

This is an educational project showing protocol fundamentals. Feel free to fork and experiment.