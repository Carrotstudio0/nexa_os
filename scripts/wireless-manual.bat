@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title "NEXA - Manual Wireless Enabler"

:: Check for Admin Rights
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo.
    echo   [!] Administrative privileges required.
    echo       Attempting to re-run with Admin rights...
    echo.
    
    REM Try to elevate
    powershell -Command "Start-Process cmd -ArgumentList '/c %~dpnx0' -Verb RunAs" >nul 2>&1
    exit /b
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
echo %CYN%       NEXA Manual Wireless Enabler v1.0                  %RST%
echo %CYN%═══════════════════════════════════════════════════════════%RST%
echo.
echo %GRA%Select Method:%RST%
echo.
echo   %CYN%1%RST% %GRA%- Try netsh (recommended, most reliable)%RST%
echo   %CYN%2%RST% %GRA%- Try Windows Hotspot (modern, if available)%RST%
echo   %CYN%3%RST% %GRA%- Try both automatically%RST%
echo   %CYN%4%RST% %GRA%- Check wireless status only%RST%
echo   %CYN%5%RST% %GRA%- Disable hotspot%RST%
echo.
set /p CHOICE="%YLW%Enter choice (1-5): %RST%"

if "%CHOICE%"=="1" goto TRY_NETSH
if "%CHOICE%"=="2" goto TRY_WINRT
if "%CHOICE%"=="3" goto TRY_BOTH
if "%CHOICE%"=="4" goto CHECK_STATUS
if "%CHOICE%"=="5" goto DISABLE_HOTSPOT

echo %RED%Invalid choice%RST%
timeout /t 2 >nul
goto :EOF

:TRY_NETSH
cls
echo %GRA%Attempting netsh method...%RST%
echo.
netsh wlan set hostednetwork mode=allow ssid=NEXA_ULTIMATE_v3.1 key=nexa123456 2>&1
echo.
echo %BLU%Starting hosted network...%RST%
netsh wlan start hostednetwork 2>&1
echo.
netsh wlan show hostednetwork 2>&1
echo.
echo %GRA%Press any key to continue...%RST%
pause >nul
goto :EOF

:TRY_WINRT
cls
echo %GRA%Attempting Windows Hotspot method (WinRT)...%RST%
echo.
powershell -ExecutionPolicy Bypass -File "..\scripts\enable-hotspot.ps1"
echo.
echo %GRA%Press any key to continue...%RST%
pause >nul
goto :EOF

:TRY_BOTH
cls
echo %GRA%Attempting automatic wireless enablement...%RST%
echo.
echo %BLU%[Method 1]%RST% %GRA%Trying netsh...%RST%
netsh wlan set hostednetwork mode=allow ssid=NEXA_ULTIMATE_v3.1 key=nexa123456 >nul 2>&1
netsh wlan start hostednetwork >nul 2>&1
echo.
echo %BLU%[Method 2]%RST% %GRA%Checking status...%RST%
netsh wlan show hostednetwork 2>&1 | find "Started" >nul
if %ERRORLEVEL% equ 0 (
    echo %GRN%✓ Hotspot is active!%RST%
) else (
    echo %YLW%[INFO] netsh method in progress or not available. Checking...%RST%
    timeout /t 2 >nul
    netsh wlan show hostednetwork 2>&1
)
echo.
echo %GRA%Press any key to continue...%RST%
pause >nul
goto :EOF

:CHECK_STATUS
cls
echo %BLU%WIRELESS STATUS CHECK%RST%
echo %GRA%═══════════════════════════════════════════════════════%RST%
echo.
echo %BLU%[Interfaces]%RST%
netsh wlan show interfaces
echo.
echo %BLU%[Hosted Network Status]%RST%
netsh wlan show hostednetwork
echo.
echo %BLU%[Drivers]%RST%
netsh wlan show drivers | find "Hosted"
echo.
echo %GRA%Press any key to continue...%RST%
pause >nul
goto :EOF

:DISABLE_HOTSPOT
cls
echo %YLW%Disabling hotspot...%RST%
echo.
netsh wlan set hostednetwork mode=disallow 2>&1
echo.
echo %GRN%✓ Hotspot disabled.%RST%
echo.
echo %GRA%Press any key to continue...%RST%
pause >nul
goto :EOF
