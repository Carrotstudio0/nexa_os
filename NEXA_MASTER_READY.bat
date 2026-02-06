@echo off
:: NEXA ULTIMATE v4.0.0-PRO | ALL-IN-ONE MASTER CONTROL
:: This script handles EVERYTHING: Firewall, Port 80, Building, and Launching.
title NEXA MASTER CONTROL

echo.
echo  [NEXA] Initializing Integrated Matrix Environment...
echo  --------------------------------------------------

:: 1. Self-Elevation Check
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo  [ERROR] This script requires Administrative Privileges.
    echo  Please right-click and 'Run as Administrator'.
    pause
    exit /b
)

:: 2. Firewall & Port Optimization
echo  [1/4] Optimizing Network Ports (80, 53, 8000)...
powershell -Command "Remove-NetFirewallRule -DisplayName 'NEXA*' -ErrorAction SilentlyContinue"
powershell -Command "New-NetFirewallRule -DisplayName 'NEXA BINARY' -Direction Inbound -Program '%CD%\nexa.exe' -Action Allow -Profile Any"
powershell -Command "New-NetFirewallRule -DisplayName 'NEXA WEB' -Direction Inbound -LocalPort 80,8000,8080 -Protocol TCP -Action Allow -Profile Any"
powershell -Command "New-NetFirewallRule -DisplayName 'NEXA DNS' -Direction Inbound -LocalPort 53,5353 -Protocol UDP -Action Allow -Profile Any"

:: 3. Liberate Port 80 (Stop Windows HTTP Service)
echo  [2/4] Liberating Port 80 from System Services...
powershell -Command "Stop-Service -Name W3SVC -Force -ErrorAction SilentlyContinue"
net stop http /y >nul 2>&1

:: 4. Build System & Cleanup
echo  [3/4] Compiling Integrity Binary & Neutralizing Port Locks...
taskkill /F /IM nexa.exe /T >nul 2>&1
taskkill /F /IM nexa_gateway.exe /T >nul 2>&1
taskkill /F /IM nexa_admin.exe /T >nul 2>&1
taskkill /F /IM nexa_dns.exe /T >nul 2>&1
taskkill /F /IM nexa_dashboard.exe /T >nul 2>&1
taskkill /F /IM nexa_core_server.exe /T >nul 2>&1
taskkill /F /IM nexa_web.exe /T >nul 2>&1
taskkill /F /IM nexa_chat.exe /T >nul 2>&1
taskkill /F /IM main.exe /T >nul 2>&1
taskkill /F /IM ISCC.exe /T >nul 2>&1
go build -trimpath -o nexa.exe ./cmd/nexa/main.go

if %errorLevel% neq 0 (
    echo  [ERROR] Compilation failed. Please check Go installation.
    pause
    exit /b
)

:: 5. Launch
echo  [4/4] Launching NEXA Matrix Engine...
echo  --------------------------------------------------
echo  [SUCCESS] System is now Online!
echo.
nexa.exe
pause
