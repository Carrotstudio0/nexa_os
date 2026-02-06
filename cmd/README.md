# NEXA Legacy Services
## ⚠️ This directory contains OLD standalone binaries

**Status:** DEPRECATED - Use `nexa.exe` (Unified Core v3.1) instead

### What changed?
In v3.1, NEXA moved from **multiple separate binaries** to a **single unified executable**.

### Legacy Binaries (Do NOT use these):
```
❌ cmd/admin/main.go       → Use: nexa.exe (Admin Panel runs at :8080)
❌ cmd/chat/main.go        → Use: nexa.exe (Chat runs at :8082)
❌ cmd/dashboard/main.go   → Use: nexa.exe (Dashboard runs at :7000)
❌ cmd/dns/main.go         → Use: nexa.exe (DNS runs at :1112)
❌ cmd/gateway/main.go     → Use: nexa.exe (Gateway runs at :8000)
❌ cmd/server/main.go      → Use: nexa.exe (Server runs at :1413)
❌ cmd/web/main.go         → Use: nexa.exe (Web runs at :3000)
```

### How to use the NEW system:
```bash
# Build the unified core
go build -o bin/nexa.exe ./cmd/nexa

# Run everything at once
./bin/nexa.exe
```

### Service Endpoints (v3.1 Unified):
```
🖥️  Dashboard  : http://localhost:7000
🚪 Gateway    : http://localhost:8000
⚙️  Admin      : http://localhost:8080
💾 Storage    : http://localhost:8081
💬 Chat       : http://localhost:8082
🔍 DNS        : localhost:1112
⚡ Core       : localhost:1413
🌐 Web        : http://localhost:3000
```

### If you need to run individual services:
Each service now lives in `pkg/services/<service>/` as Go packages, NOT as standalone binaries.

To integrate or extend a service, edit the corresponding file in:
```
pkg/services/
├── admin/
├── chat/
├── dashboard/
├── dns/
├── gateway/
├── server/
├── storage/
└── web/
```

### Migration Guide:
| OLD | NEW |
|:---|:---|
| `cmd/admin/main.go` | `pkg/services/admin/` |
| `cmd/chat/main.go` | `pkg/services/chat/` |
| Standalone CLI | Unified `nexa.exe` |
| 7 separate processes | 1 unified process |

---

**Version:** v3.1 (Unified Core)  
**Date:** February 2026  
**Recommendation:** Never run the legacy binaries. Always use `nexa.exe`
