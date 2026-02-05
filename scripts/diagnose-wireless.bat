@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title "NEXA Wireless Diagnostic Tool"

:: Check for Admin Rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo.
    echo   [!] CRITICAL: Administrative privileges required.
    echo       Please right-click and 'Run as Administrator'.
    echo.
    pause
    exit /b 1
)

for /f %%a in ('echo prompt $E ^| cmd') do set "ESC=%%a"
set "BLU=%ESC%[94m"
set "GRA=%ESC%[90m"
set "GRN=%ESC%[92m"
set "YLW=%ESC%[93m"
set "RED=%ESC%[91m"
set "CYN=%ESC%[96m"
set "RST=%ESC%[0m"

cls
echo.
echo %CYN%═══════════════════════════════════════════════════════════%RST%
echo %CYN%       NEXA Wireless Diagnostic Tool v1.0                  %RST%
echo %CYN%═══════════════════════════════════════════════════════════%RST%
echo.

echo %BLU%[1]%RST% %GRA%Checking Wireless Adapters...%RST%
echo %GRA%─────────────────────────────────────────────────%RST%
netsh wlan show interfaces
echo.

echo %BLU%[2]%RST% %GRA%Checking Wireless Drivers...%RST%
echo %GRA%─────────────────────────────────────────────────%RST%
netsh wlan show drivers
echo.

echo %BLU%[3]%RST% %GRA%Checking Hosted Network Support...%RST%
echo %GRA%─────────────────────────────────────────────────%RST%
netsh wlan show hostednetwork
echo.

echo %BLU%[4]%RST% %GRA%Checking Network Services...%RST%
echo %GRA%─────────────────────────────────────────────────%RST%
sc query "wlansvc" | find "STATE"
sc query "dot3svc" | find "STATE"
sc query "netsvc" | find "STATE"
echo.

echo %BLU%[RECOMMENDATIONS]%RST% %GRN%Fix Steps:%RST%
echo %GRA%─────────────────────────────────────────────────%RST%
echo.
echo %YLW%If "Hosted network supported" shows "No":%RST%
echo  1. Update your wireless card drivers (from manufacturer website)
echo  2. Enable wireless adapter in Device Manager
echo  3. Restart the computer after driver update
echo.
echo %YLW%If wireless service is disabled:%RST%
echo  1. Open Services (services.msc)
echo  2. Find "WLAN AutoConfig" and set to Automatic
echo  3. Click Start
echo.
echo %YLW%To manually enable Hosted Network:%RST%
echo  Run this command in Admin PowerShell:
echo  ^> netsh wlan set hostednetwork mode=allow ssid=NEXA_ULTIMATE_v3.1 key=nexa123456
echo  ^> netsh wlan start hostednetwork
echo.
echo %GRA%═══════════════════════════════════════════════════════════%RST%
echo.

pause
exit /b 0
