@echo off
title DISABLE FIREWALL (TEMPORARY TEST)
echo ==========================================
echo WARNING: TURNING OFF FIREWALL FOR TESTING
echo ==========================================

:: Check Admin
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo FAIL: Right-click and "Run as Administrator"
    pause
    exit
)

echo 1. Turning off ALL Firewall Profiles...
netsh advfirewall set allprofiles state off

echo.
echo 2. Current IP Address:
ipconfig | findstr "IPv4"

echo.
echo ==========================================
echo FIREWALL IS OFF. TRY THE PHONE NOW.
echo ==========================================
echo.
echo If it works: The problem was the Firewall.
echo If it fails: The problem is the Network/IP.
echo.
pause
