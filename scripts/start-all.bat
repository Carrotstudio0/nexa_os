@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title "NEXA ULTIMATE v3.1 - Unified Core Command Center"

:: Admin Check
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [!] Administrative privileges required.
    pause
    exit /b 1
)

:: Colors
for /f %%a in ('echo prompt $E ^| cmd') do set "ESC=%%a"
set "GRA=%ESC%[90m"
set "GRN=%ESC%[92m"
set "RED=%ESC%[91m"
set "CYN=%ESC%[96m"
set "YLW=%ESC%[93m"
set "MAG=%ESC%[95m"
set "RST=%ESC%[0m"

cls
echo.
echo %CYN%    _   _______  _____     __  ____  __________ 
echo %CYN%   / \ / /  __/ /_  _/    / / / / / /_  __/ __ \
echo %CYN%  /   / /  __/   / /     / /_/ / /   / / / /_/ /
echo %CYN% /_/ \_/\___/   /_/      \____/_/   /_/  \____/  %GRA%v3.1 - Unified%RST%
echo.

cd /d "%~dp0.."

:: --- Network Layer Configuration ---

REM --- Hotspot is disabled by default for التشغيل التلقائي. لتفعيل الهوتسبوت تلقائياً غيّر السطر التالي إلى Y ---
set "HOTSPOT=N"
REM --- إذا أردت تفعيل الهوتسبوت تلقائياً، غيّر السطر أعلاه إلى Y ---
if /i "!HOTSPOT!" neq "Y" goto :START_CORE

echo.
echo %CYN%[HOTSPOT]%RST% %GRA%Initializing wireless transmission...%RST%

if exist "scripts\enable-hotspot.ps1" (
    powershell -ExecutionPolicy Bypass -File "scripts\enable-hotspot.ps1"
    if !errorlevel! equ 0 (
        echo   %GRN%✓ WiFi Hotspot Enabled%RST%
    ) else (
        echo   %RED%✖ Hotspot setup failed (continuing with local network)%RST%
        timeout /t 2 >nul
    )
) else (
    echo   %RED%✖ Hotspot script not found in scripts\%RST%
    echo   %GRA%Continuing without hotspot...%RST%
    timeout /t 2 >nul
)
echo.


:START_CORE
:: --- Unified Service Matrix Initialization ---
echo %CYN%[CORE]%RST% %GRN%Launching Unified Nexa System...%RST%
echo.

REM --- Build automatically if nexa.exe missing ---
if not exist "bin\nexa.exe" (
    echo   %YLW%[!] bin\nexa.exe not found. Compiling automatically...%RST%
    cd /d "%~dp0.."
    call scripts\build.bat
    cd /d "%~dp0"
)

cd /d "%~dp0.."
start "NEXA ULTIMATE v3.1 - CORE" bin\nexa.exe

:: Get Local IP for display
for /f "tokens=4 delims= " %%i in ('route print ^| findstr 0.0.0.0 ^| findstr /V "127.0.0.1" ^| findstr /V "::"') do set "MY_IP=%%i"
if "!MY_IP!"=="" set "MY_IP=localhost"

:: Wait for services to start
timeout /t 3 >nul

cls
echo.
echo %CYN%╔═══════════════════════════════════════════════════════════╗%RST%
echo %CYN%║         NEXA SYSTEM FULLY OPERATIONAL                    ║%RST%
echo %CYN%╚═══════════════════════════════════════════════════════════╝%RST%
echo.
echo %GRA%Primary Services Online:%RST%
echo.
echo   %GRN%✓%RST% %CYN%Dashboard  %RST%%GRA%: %RST%%CYN%http://!MY_IP!:7000%RST%   %GRA%(Main Hub)%RST%
echo   %GRN%✓%RST% %CYN%Gateway    %RST%%GRA%: %RST%%CYN%http://!MY_IP!:8000%RST%   %GRA%(Routing)%RST%
echo   %GRN%✓%RST% %CYN%Admin Panel%RST%%GRA%: %RST%%CYN%http://!MY_IP!:8080%RST%   %GRA%(Management)%RST%
echo.
echo %GRA%Storage ^& Communication:%RST%
echo.
echo   %GRN%✓%RST% %CYN%Storage    %RST%%GRA%: %RST%%CYN%http://!MY_IP!:8081%RST%   %GRA%(Files + Vault)%RST%
echo   %GRN%✓%RST% %CYN%Chat       %RST%%GRA%: %RST%%CYN%http://!MY_IP!:8082%RST%   %GRA%(Messaging)%RST%
echo   %GRN%✓%RST% %CYN%Web        %RST%%GRA%: %RST%%CYN%http://!MY_IP!:3000%RST%   %GRA%(New Service)%RST%
echo.
echo %GRA%Backend Services:%RST%
echo.
echo   %GRN%✓%RST% %CYN%Core Server%RST%%GRA%: %RST%%CYN%localhost:1413%RST%  %GRA%(Ledger + Blockchain)%RST%
echo   %GRN%✓%RST% %CYN%DNS Server %RST%%GRA%: %RST%%CYN%localhost:1112%RST%  %GRA%(Name Resolution)%RST%
echo.
echo %GRA%══════════════════════════════════════════════════════════%RST%
echo.
REM --- Open browser automatically to Dashboard ---
echo %YLW%[INFO]%RST% Browser opening to Dashboard. Keep this window open.
timeout /t 2 >nul
start "" http://!MY_IP!:7000
echo %YLW%[INFO]%RST% All 8 services running in unified nexa.exe process.
echo.
echo %RED%Press Ctrl+C to shutdown all services safely.%RST%
echo %RED%Or close this window to terminate the system.%RST%
echo.

:: Keep window open - monitor the core process
:MONITOR
timeout /t 5 >nul
tasklist /FI "IMAGENAME eq nexa.exe" 2>NUL | find /I /N "nexa.exe">NUL
if %ERRORLEVEL% == 0 (
    goto MONITOR
) else (
    echo %GRA%[NOTICE] Core process ended.%RST%
)

echo.
echo %GRA%[SHUTDOWN] System offline.%RST%
timeout /t 2 >nul
exit /b
