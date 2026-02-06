# Nexa Universal Server - Docker Guide

## 🚀 Quick Start (Production Deployment)

### 1. Build and Run with Docker Compose
```bash
docker-compose up -d
```

This will:
- Build the Nexa OS Docker image
- Start the server in the background
- Expose all necessary ports
- Mount `./data` and `./sites` directories for persistence

### 2. Access Your Services
- **Dashboard (Intelligence Hub)**: http://YOUR_SERVER_IP:7000
- **Gateway**: http://YOUR_SERVER_IP:8000
- **Admin Panel**: http://YOUR_SERVER_IP:8080
- **Storage**: http://YOUR_SERVER_IP:8081
- **Chat**: http://YOUR_SERVER_IP:8082

### 3. View Logs
```bash
docker-compose logs -f nexa
```

### 4. Stop the Server
```bash
docker-compose down
```

## 🛠 Manual Docker Build

If you prefer manual control:

```bash
# Build the image
docker build -t nexa-os:latest .

# Run the container
docker run -d \
  --name nexa_server \
  -p 80:8000 \
  -p 7000:7000 \
  -p 8000:8000 \
  -p 8080:8080 \
  -p 8081:8081 \
  -p 8082:8082 \
  -p 53:53/udp \
  -v $(pwd)/data:/root/data \
  -v $(pwd)/sites:/root/sites \
  nexa-os:latest
```

## 📦 Environment Variables

You can customize ports via environment variables:

```yaml
environment:
  - NEXA_GATEWAY_PORT=8000
  - NEXA_DASHBOARD_PORT=7000
  - NEXA_ADMIN_PORT=8080
  - NEXA_STORAGE_PORT=8081
  - NEXA_CHAT_PORT=8082
  - NEXA_DNS_PORT=1112
```

## 🌍 Deploy Anywhere

This Docker image works on:
- ✅ Linux VPS (AWS, DigitalOcean, Linode, etc.)
- ✅ Raspberry Pi (ARM64)
- ✅ Windows (via Docker Desktop)
- ✅ macOS (via Docker Desktop)
- ✅ Kubernetes clusters

## 📁 Data Persistence

All data is stored in:
- `./data/` - Application data
- `./sites/` - Hosted websites/projects

These directories are automatically mounted as volumes.

## 🔧 Configuration

Edit `config.yaml` before building to customize:
- System name and version
- Service ports
- Network settings
- File paths

## 🐳 Multi-Platform Build (Optional)

To build for multiple architectures:

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t nexa-os:latest .
```

## 📝 Notes

- The container runs as root by default (required for port 53/80)
- For production, consider using a reverse proxy (nginx/traefik)
- SSL/TLS certificates can be mounted via volumes
