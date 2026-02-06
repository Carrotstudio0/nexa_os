# 🚀 QUICK START GUIDE - NEXA OS v4.0.0-PRO

**Last Updated**: February 6, 2026 | **Version**: Production Ready

---

## ⚡ 60-SECOND STARTUP

### 1️⃣ Navigate to bin folder
```cmd
cd bin
```

### 2️⃣ Run the system
```cmd
nexa.exe
```

### 3️⃣ Open dashboard
Go to: **http://localhost:7000**

✅ **DONE!** System is running.

---

## 📍 SERVICE LOCATIONS

After startup, access any service directly:

| Service | URL | Port | Purpose |
|---------|-----|------|---------|
| Dashboard | http://localhost:7000 | 7000 | Main hub & control center |
| Gateway | http://localhost:8000 | 8000 | API proxy & routing |
| Admin | http://localhost:8080 | 8080 | System administration |
| Storage | http://localhost:8081 | 8081 | File management |
| Chat | http://localhost:8082 | 8082 | Messaging system |
| Web | http://localhost:3000 | 3000 | Web service |
| DNS | localhost:53 | 53 | DNS resolution |
| Core | localhost:1413 | 1413 | Backend server |

---

## 🔐 DEFAULT CREDENTIALS

**Username**: `admin`  
**Password**: `admin123`  

⚠️ **Change these immediately in production!**

---

## ✅ VERIFY INSTALLATION

### Check if services are running
```cmd
:: From another terminal
curl http://localhost:7000/health
curl http://localhost:8000/health
curl http://localhost:8080/health
```

### Expected response
```json
{"status":"healthy","service":"dashboard","timestamp":1707123456}
```

---

## 🛠️ COMMON TASKS

### Rebuild from Source
```cmd
cd ..
go build -o bin/nexa.exe .\cmd\nexa
.\bin\nexa.exe
```

### Check System Status
```
Visit: http://localhost:7000
Look at: Dashboard → System Monitor
```

### View Logs
```
Check console output while nexa.exe is running
Logs appear in real-time with color codes
```

### Stop the System
```
Press: Ctrl+C in the terminal window
System will gracefully shutdown
```

---

## 🔧 TROUBLESHOOTING

### Port Already in Use
**Error**: Address already in use  
**Solution**: 
```cmd
:: Find process using the port
netstat -ano | findstr :7000

:: Stop the process
taskkill /PID [PID] /F

:: Or change the port in config.yaml
```

### Permission Denied
**Error**: Access denied when creating rules  
**Solution**: 
```
Run as Administrator:
- Right-click Command Prompt
- Select "Run as Administrator"
- Then run bin\nexa.exe
```

### Services Not Responding
**Error**: Cannot connect to localhost:8000  
**Solution**:
```
1. Check if nexa.exe is still running
2. Look for error messages in console
3. Check firewall isn't blocking ports
4. Restart the service
```

### Go Not Found
**Error**: 'go' is not recognized  
**Solution**:
```
Install Go from https://golang.org/dl/
Or copy pre-built nexa.exe from bin/
```

---

## 📊 SYSTEM REQUIREMENTS

### Minimum
- Windows 10 or Linux Ubuntu 20.04+
- 256 MB RAM free
- 100 MB disk space
- No external dependencies (standalone executable)

### Recommended
- Windows 11 or Linux Ubuntu 22.04+
- 1 GB RAM
- 500 MB disk space
- Administrator/Root access (for network features)

---

## 📁 FOLDER STRUCTURE

```
nexa_os/
├── bin/
│   ├── nexa.exe              ← RUN THIS
│   ├── config.yaml           ← Configuration
│   ├── users.json            ← User database
│   ├── ledger.json           ← Blockchain
│   └── ...other files
├── pkg/                       ← Source code (services)
├── cmd/                       ← Command line apps
├── config/                    ← Config templates
├── storage/                   ← File storage
├── data/                      ← System data
└── docs/                      ← Documentation
```

---

## 🔑 IMPORTANT FILES

| File | Purpose | Edit? |
|------|---------|-------|
| `nexa.exe` | Main executable | ❌ No |
| `config.yaml` | System config | ✅ Yes |
| `users.json` | User credentials | ✅ Yes |
| `dns_records.json` | DNS entries | ✅ Yes |
| `ledger.json` | Blockchain | ⚠️ Careful |

---

## 🔄 STARTUP SEQUENCE

When you run `nexa.exe`, the system:

1. ✅ Loads configuration (config.yaml + defaults)
2. ✅ Initializes network manager
3. ✅ Starts governance system
4. ✅ Loads blockchain ledger
5. ✅ Initializes authentication
6. ✅ Launches 8 integrated services in parallel
7. ✅ Opens dashboard browser window
8. ✅ Reports ready status

**Total time**: ~3 seconds

---

## 🚨 FIRST-TIME SETUP

### If this is your first run:

1. **Verify startup**
   ```
   Check console for: "Matrix is fully operational"
   ```

2. **Open dashboard**
   ```
   Browser should auto-open
   If not: http://localhost:7000
   ```

3. **Set admin password**
   ```
   Dashboard → Admin → Change Password
   ```

4. **Create additional users** (optional)
   ```
   Admin Panel → Users → Add User
   ```

5. **Configure profiles**
   ```
   Settings → Preferences → Your Organization
   ```

---

## 💾 DATA PERSISTENCE

All system data is automatically saved:

- **Users**: `users.json`
- **DNS Records**: `dns_records.json`
- **Blockchain**: `ledger.json`
- **Configuration**: `config.yaml`
- **Files**: `storage/` directory

🔄 *Automatic backup recommended for production!*

---

## 🌐 NETWORK MODE

### Local Network Access
To access from another computer on your network:

Replace `localhost` with your computer's IP:
```
http://192.168.1.100:7000
(substitute your actual IP)
```

Find your IP:
```cmd
ipconfig
:: Look for "IPv4 Address" under your network adapter
```

---

## 📞 GETTING HELP

### Check Documentation
- 📖 [PRODUCTION_READY.md](../PRODUCTION_READY.md) - Detailed guide
- 📖 [AUDIT_REPORT.md](../AUDIT_REPORT.md) - Technical details
- 📖 [readme.md](../readme.md) - Full documentation

### View Status
Dashboard includes:
- System health
- Service status
- Network topology
- Performance metrics

### Check Logs
- Console output shows all activity
- Real-time status messages
- Error alerts and warnings

---

## 🎯 NEXT STEPS

1. **Try dashboard features**
   - Navigate all sections
   - Test file upload
   - Send test messages

2. **Invite other users**
   - Create new accounts
   - Assign roles
   - Test access control

3. **Set up backups**
   - Regular ledger snapshots
   - User data export
   - Configuration backup

4. **Plan expansion**
   - Add more devices
   - Configure DNS
   - Expand storage

---

## ✨ TIPS & TRICKS

### Keyboard Shortcuts
```
Ctrl + C  →  Stop system gracefully
Ctrl + L  →  Clear console (in some terminals)
Tab       →  Autocomplete in API endpoints
```

### Configuration Tweaks
Edit `config.yaml` to:
```yaml
# Change ports
services:
  dashboard:
    port: 7001

# Adjust timeouts
server:
  timeout: 60
```

### Performance Boost
1. Close unused services in dashboard
2. Disable metrics for unused services
3. Reduce polling intervals

---

## 🔐 SECURITY REMINDER

- **Change default password immediately**
- Keep `users.json` file secure
- Use HTTPS in production (configure certificates)
- Regularly backup `ledger.json`
- Rotate user credentials periodically

---

## 📊 MONITORING

Check system health:
```
Dashboard → Monitor → System Status
Or visit: http://localhost:7000/health
```

Key metrics:
- CPU usage
- Memory consumption
- Active connections
- Request rate
- Error count

---

## 🎓 LEARNING PATH

1. **Understanding** (15 min)
   - Read: [readme.md](../readme.md)
   - Understand: System architecture

2. **Exploration** (30 min)
   - Run: `nexa.exe`
   - Visit: http://localhost:7000
   - Click: Each menu item

3. **Configuration** (20 min)
   - Edit: config.yaml
   - Create users
   - Test endpoints

4. **Customization** (variable)
   - Modify services
   - Add features
   - Integrate systems

---

## 🎉 YOU'RE READY!

The system is now running and ready for:
- ✅ Development
- ✅ Testing
- ✅ Production deployment
- ✅ Scaling

**Enjoy NEXA OS v4.0.0-PRO!**

---

### Quick Support
1. Error? → Check console output
2. Stuck? → Read PRODUCTION_READY.md
3. Question? → Check readme.md
4. Problem? → Review AUDIT_REPORT.md

**Status**: ✅ System Ready | Version: v4.0.0-PRO | Date: Feb 6, 2026
