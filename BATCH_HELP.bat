@echo off
REM This file displays help information about NEXA batch scripts
setlocal enabledelayedexpansion

for /f %%a in ('echo prompt $E ^| cmd') do set "ESC=%%a"
set "CYN=%ESC%[96m"
set "GRN=%ESC%[92m"
set "GRA=%ESC%[90m"
set "YLW=%ESC%[93m"
set "RST=%ESC%[0m"

cls
echo.
echo %CYN%╔═══════════════════════════════════════════════════════════╗%RST%
echo %CYN%║   NEXA v3.1 BATCH FILES GUIDE                            ║%RST%
echo %CYN%╚═══════════════════════════════════════════════════════════╝%RST%
echo.

echo %GRN%Main Entry Points:%RST%
echo.
echo   %CYN%NEXA.bat%RST%
echo     %GRA%Purpose:%RST% Main launcher menu for the entire system
echo     %GRA%Usage:%RST%   Double-click or run from command line
echo     %GRA%Features:%RST% Detects dev/prod mode, builds or launches
echo.
echo   %CYN%Building/%RST%
echo     %GRA%├─ scripts\build.bat%RST%
echo     %GRA%│  Purpose: Compile the unified nexa.exe from source
echo     %GRA%│  Usage:   Called by NEXA.bat or run manually
echo     %GRA%│  Time:    30-60 seconds (first time)
echo     %GRA%│
echo     %GRA%└─ Result: bin\nexa.exe (all 8 services included)
echo.
echo   %CYN%Starting system:%RST%
echo     %GRA%└─ scripts\start-all.bat%RST%
echo     %GRA%   Purpose: Unified launcher for the entire system
echo     %GRA%   Usage:   Always use this file to start NEXA
echo     %GRA%   Features: Auto-build, browser open, all-in-one, no prompts
echo.
echo   %CYN%Advanced:%RST%
echo     %GRA%├─ scripts\troubleshoot.bat%RST%
echo     %GRA%│  Purpose: Diagnose and fix issues
echo     %GRA%│  Features: Check processes, ports, logs, rebuild
echo     %GRA%│
echo     %GRA%├─ scripts\enable-hotspot.ps1%RST%
echo     %GRA%│  Purpose: Setup WiFi hotspot (optional)
echo     %GRA%│  Usage:   Called by start-all.bat
echo     %GRA%│
echo     %GRA%└─ scripts\detect-hotspot.ps1%RST%
echo        Purpose: Detect available network interfaces
echo.

echo %GRN%Services Included in nexa.exe:%RST%
echo.
echo   %CYN%Primary Interfaces:%RST%
echo     %GRA%✓ Dashboard   - port 7000  (Main UI hub)%RST%
echo     %GRA%✓ Gateway     - port 8000  (Request routing)%RST%
echo     %GRA%✓ Admin Panel - port 8080  (System management)%RST%
echo.
echo   %CYN%Storage ^& Communication:%RST%
echo     %GRA%✓ Storage     - port 8081  (File + Vault)%RST%
echo     %GRA%✓ Chat        - port 8082  (Messaging)%RST%
echo     %GRA%✓ Web         - port 3000  (Web service - NEW)%RST%
echo.
echo   %CYN%Backend:%RST%
echo     %GRA%✓ Core Server - port 1413  (Ledger + Blockchain)%RST%
echo     %GRA%✓ DNS         - port 1112  (Name resolution)%RST%
echo.

echo %GRN%Quick Start Workflow:%RST%
echo.
echo   %CYN%1. First time setup:%RST%
echo      %GRA%$ NEXA.bat %RST%
echo      %GRA%Select option 1: Build ^& Launch%RST%
echo.
echo   %CYN%2. Subsequent starts:%RST%
echo      %GRA%$ NEXA.bat%RST%
echo      %GRA%Select option 2: Launch (use existing binaries)%RST%
echo.
echo   %CYN%3. Direct launch (if built):%RST%
echo      %GRA%$ bin\start-all.bat%RST%
echo.

echo %GRN%Troubleshooting:%RST%
echo.
echo   %CYN%If system won't start:%RST%
echo     %GRA%$ scripts\troubleshoot.bat %RST%
echo     %GRA%Select options 3 (clean build) or 6 (debug mode)%RST%
echo.
echo   %CYN%If ports are in use:%RST%
echo     %GRA%$ scripts\troubleshoot.bat %RST%
echo     %GRA%Select option 4 (check port availability)%RST%
echo.

echo %GRN%File Changes in v3.1:%RST%
echo.
echo   %CYN%What's NEW:%RST%
echo     %GRA%✓ Unified Core (1 binary: nexa.exe)%RST%
echo     %GRA%✓ Web Service included (port 3000)%RST%
echo     %GRA%✓ Better error handling in all .bat files%RST%
echo     %GRA%✓ Improved ledger persistence%RST%
echo.
echo   %CYN%What's REMOVED:%RST%
echo     %GRA%✗ Multiple .exe files (dns.exe, server.exe, etc.)%RST%
echo     %GRA%✗ Complex service coordination (now unified)%RST%
echo.

echo %GRN%Important Notes:%RST%
echo.
echo   %GRA%• Admin privileges required for all operations%RST%
echo   %GRA%• First build takes 30-60 seconds (subsequent: fastest)%RST%
echo   %GRA%• Keep the console window open while system is running%RST%
echo   %GRA%• Press Ctrl+C to gracefully shutdown all services%RST%
echo.

pause
