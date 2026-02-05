@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title "NEXA ULTIMATE | Command Center v3.1"

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
echo %CYN% /_/ \_/\___/   /_/      \____/_/   /_/  \____/  %GRA%v3.1%RST%
echo.

cd /d "%~dp0"

:: --- Network Layer Configuration ---
set "HOTSPOT=N"
echo %CYN%[NETWORK]%RST% %GRA%Configuring connection layer...%RST%
set /p HOTSPOT="%YLW%[?] Activate WiFi hotspot matrix? (Y/N): %RST%"

if /i "!HOTSPOT!" neq "Y" goto :START_MATRIX

echo.
echo %CYN%[HOTSPOT]%RST% %GRA%Initializing wireless transmission...%RST%

rem Check for scripts
if exist "..\scripts\enable-hotspot.ps1" (
    powershell -ExecutionPolicy Bypass -File "..\scripts\enable-hotspot.ps1"
) else if exist "..\scripts\detect-hotspot.ps1" (
    powershell -ExecutionPolicy Bypass -File "..\scripts\detect-hotspot.ps1" -TimeoutSeconds 20
) else (
    echo   %RED%✖ Hotspot scripts and resources not found in \scripts%RST%
)

if !errorlevel! equ 0 (
    echo   %GRN%✓ Hotspot Layer Enabled%RST%
) else (
    echo   %RED%✖ Hotspot initialization failed. Continuing with local network.%RST%
    pause
)
echo.

:START_MATRIX
:: --- Unified Service Matrix Initialization ---
echo %CYN%[MATRIX]%RST% %GRA%Launching Unified Nexa Core...%RST%

if exist nexa.exe (
    echo   %GRN%✓%RST% Starting Nexa Multi-Service Engine
    :: We run it in a new window but just ONE window
    start "NEXA ULTIMATE | CORE" nexa.exe
) else (
    echo   %RED%✖ nexa.exe missing - please run BUILD.bat%RST%
    pause
    exit /b
)

echo.

 :: Get Local IP for display
for /f "tokens=4 delims= " %%i in ('route print ^| findstr 0.0.0.0 ^| findstr /V "127.0.0.1" ^| findstr /V "::"') do set "MY_IP=%%i"
if "!MY_IP!"=="" set "MY_IP=localhost"

echo   %GRA%══════════════════════════════════════════════════════════%RST%
echo   %GRN%  SYSTEM FULLY OPERATIONAL%RST%
echo.
echo   %GRA%  🌐 Dashboard  : %RST%%CYN%http://!MY_IP!:7000%RST%
echo   %GRA%  🚪 Gateway    : %RST%%CYN%http://!MY_IP!:8000%RST%
echo.

:: Open Dashboard
timeout /t 2 >nul
start http://localhost:7000

echo %RED%  [WARNING] Keep this window open - it supervises all matrix services.%RST%
echo %GRA%  Press any key to execute full shutdown sequence...%RST%
pause >nul

echo.
echo %MAG%[SHUTDOWN] Terminating active Matrix Core...%RST%
taskkill /F /IM nexa.exe >nul 2>&1

echo %GRN%[DONE] All systems safe and offline.%RST%
timeout /t 2 >nul
exit /b
