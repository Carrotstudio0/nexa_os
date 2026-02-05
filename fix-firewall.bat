@echo off
title Opening Nexa Ports (Admin Required)
echo Requesting Admin Privileges...

:: Check for admin rights
net session >nul 2>&1
if %errorLevel% == 0 (
    echo Success: Adding Firewall Rules...
) else (
    echo Please right-click and "Run as Administrator"
    pause
    exit
)

echo Opening Port 7000 (Dashboard)...
netsh advfirewall firewall add rule name="Nexa Dashboard" dir=in action=allow protocol=TCP localport=7000 profile=any

echo Opening Port 8081 (File Manager)...
netsh advfirewall firewall add rule name="Nexa FileManager" dir=in action=allow protocol=TCP localport=8081 profile=any

echo Opening Port 8000 (Gateway)...
netsh advfirewall firewall add rule name="Nexa Gateway" dir=in action=allow protocol=TCP localport=8000 profile=any

echo Opening Port 8080 (Admin)...
netsh advfirewall firewall add rule name="Nexa Admin" dir=in action=allow protocol=TCP localport=8080 profile=any

echo Opening Port 9000 (Server)...
netsh advfirewall firewall add rule name="Nexa Server" dir=in action=allow protocol=TCP localport=9000 profile=any

echo Opening Port 53 (DNS)...
netsh advfirewall firewall add rule name="Nexa DNS" dir=in action=allow protocol=UDP localport=53 profile=any

echo.
echo ==================================================
echo ✅ All Ports Open! Try the phone again now.
echo ==================================================
pause
