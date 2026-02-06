@echo off
:: NEXA OS ENVIRONMENT SETUP
:: This script configures the local machine for Nexa OS operation.
:: It handles Firewall rules and Port occupancy.

title NEXA ENVIRONMENT SETUP
echo.
echo  [NEXA] Configuring System Environment...
echo  --------------------------------------------------

:: 1. Self-Elevation Check (Redundant if run from installer but safe)
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [ERROR] Admin privileges required.
    exit /b 1
)

:: 2. Firewall Rules
echo [1/2] Configuring Matrix Firewall Rules...
powershell -Command "Remove-NetFirewallRule -DisplayName 'NEXA*' -ErrorAction SilentlyContinue"
powershell -Command "New-NetFirewallRule -DisplayName 'NEXA BINARY' -Direction Inbound -Program '%~dp0nexa.exe' -Action Allow -Profile Any"
powershell -Command "New-NetFirewallRule -DisplayName 'NEXA WEB' -Direction Inbound -LocalPort 80,8000,8080,7000 -Protocol TCP -Action Allow -Profile Any"
powershell -Command "New-NetFirewallRule -DisplayName 'NEXA DNS' -Direction Inbound -LocalPort 53,5353 -Protocol UDP -Action Allow -Profile Any"

:: 3. Port 80 Liberation (Optional/Aggressive)
echo [2/2] Liberating Port 80 for Gateway...
powershell -Command "Stop-Service -Name W3SVC -Force -ErrorAction SilentlyContinue"
net stop http /y >nul 2>&1

echo.
echo [SUCCESS] Environment Ready!
timeout /t 3 >nul
exit /b 0
