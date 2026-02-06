# Nexa OS Transformation Plan: Universal Server

## 1. Executive Summary
The current Nexa OS project is a powerful concept: a unified server running multiple services (Gateway, DNS, Dashboard, etc.) from a single binary. However, it is currently **heavily coupled to Windows**, making it impossible to run on Linux servers (VPS), Raspberry Pi, or macOS without major errors.

To achieve the goal of a "powerful integrated server that works on any device," we must decouple the core logic from the operating system and containerize the application.

## 2. Analysis of Current State

### Strengths
- **Unified Architecture**: `cmd/nexa/main.go` correctly orchestrates multiple services in one process.
- **Service Modularity**: `pkg/services` separates logic reasonably well.
- **No Heavy Dependencies**: It uses Go's standard library mostly, which is great for portability.

### Weaknesses & Blockers
1.  **Windows Hardcoding**:
    -   `setup_windows.go` and `network.go` contain direct PowerShell commands and `netsh` calls.
    -   Hardcoded paths like `C:\Windows\System32\drivers\etc\hosts`.
    -   Host file modification requires Admin/Root privileges, which causes failures in non-admin environments.
2.  **Configuration Rigidity**:
    -   Ports and settings are hardcoded in `config.go` or scattered in `main.go`.
    -   No support for a declarative config file (e.g., `config.yaml`).
3.  **Lack of Containerization**:
    -   No `Dockerfile` exists. This is the #1 requirement for "running seamlessly on any device."

## 3. The Transformation Roadmap

### Phase 1: Cross-Platform Core (The Foundation)
**Goal:** Make the Go code compile and run on Linux/macOS without errors.
-   [x] **Refactor `pkg/utils`**: Split `network.go` and `setup_windows.go` into:
    -   `platform_windows.go`: Existing Windows logic (PowerShell, netsh).
    -   `platform_unix.go`: Linux/macOS implementations (or no-op stubs).
    -   Use Build Tags (`//go:build windows`) to separate them.
-   [x] **Safe Startup**: Ensure the app doesn't crash if it can't bind port 53 (DNS) or 80 (Gateway) due to permission issues.

### Phase 2: Configuration Architecture
**Goal:** Allow users to change ports/settings without recompiling.
-   [x] **Central Config**: Implement a `config.yaml` loader.
-   [x] **Environment Variables**: Support `.env` files for easy deployment (via `config.go` logic).

### Phase 3: Dockerization (The "Run Anywhere" Solution)
**Goal:** One-click deploy on any server (AWS, DigitalOcean, Raspberry Pi).
-   [x] **Create `Dockerfile`**: A multi-stage build producing a tiny Alpine Linux image.
-   [x] **Create `docker-compose.yml`**: To define networks, volumes (for `./data`), and ports.

### Phase 5: Analytics & Visualization
**Goal:** Track real-time usage across the network and visualize it.
- [x] **New `pkg/analytics`**: Centralized tracking for sessions, actions, and files.
- [x] **Real-time Engine**: WebSocket-driven event streaming.
- [x] **Professional Dashboard**: Premium UI for monitoring visits, devices, and file activities.
- [x] **Service Integration**: Tracking integrated into Gateway, Storage, and Chat.

### Phase 4: Reliability & Professionalism
-   [ ] **Structured Logging**: Replace `fmt.Println` with a professional logger (Zap/Logrus) for better debugging.
-   [ ] **Graceful Shutdown**: Ensure data is saved and connections closed when the server stops.

## 5. Current Status Summary

✅ **Phase 1**: Complete - Cross-platform code with build tags  
✅ **Phase 2**: Complete - Dynamic configuration via `config.yaml`  
✅ **Phase 3**: Complete - Docker ready with `Dockerfile` and `docker-compose.yml`  
✅ **Phase 4**: In Progress - Advanced reliability features  
✅ **Phase 5**: Complete - Real-time Analytics and Monitoring Dashboard  

The project is **v4.0.0-PRO compatible**, featuring full lifecycle tracking and enterprise-grade visualization.
