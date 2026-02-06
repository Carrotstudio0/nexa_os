@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title NEXA Troubleshooter v3.1 - Unified Core

for /f %%a in ('echo prompt $E ^| cmd') do set "ESC=%%a"
set "RED=%ESC%[91m"
set "GRN=%ESC%[92m"
set "YLW=%ESC%[93m"
set "CYN=%ESC%[96m"
set "GRA=%ESC%[90m"
set "RST=%ESC%[0m"

cls
echo.
echo %CYN%╔══════════════════════════════════════════════════════╗%RST%
echo %CYN%║       NEXA Unified Core Troubleshooting              ║%RST%
echo %CYN%╚══════════════════════════════════════════════════════╝%RST%
echo.
echo %GRA%Diagnostics for v3.1 (8 services in nexa.exe):%RST%
echo.
echo   %CYN%1%RST% %GRA%- Check if nexa.exe is running%RST%
echo   %CYN%2%RST% %GRA%- Kill nexa.exe process%RST%
echo   %CYN%3%RST% %GRA%- Clean build (rebuild nexa.exe)%RST%
echo   %CYN%4%RST% %GRA%- Test port availability%RST%
echo   %CYN%5%RST% %GRA%- View system logs%RST%
echo   %CYN%6%RST% %GRA%- Run nexa.exe in console (debug)%RST%
echo   %CYN%7%RST% %GRA%- Verify Go installation%RST%
echo   %CYN%8%RST% %GRA%- Check ledger.json integrity%RST%
echo   %CYN%0%RST% %GRA%- Back to main menu%RST%
echo.
set /p CHOICE="%YLW%Enter choice (0-8): %RST%"

if "%CHOICE%"=="1" goto CHECK_SERVICES
if "%CHOICE%"=="2" goto KILL_SERVICES
if "%CHOICE%"=="3" goto CLEAN_BUILD
if "%CHOICE%"=="4" goto CHECK_PORTS
if "%CHOICE%"=="5" goto VIEW_LOGS
if "%CHOICE%"=="6" goto RUN_DEBUG
if "%CHOICE%"=="7" goto CHECK_GO
if "%CHOICE%"=="8" goto CHECK_LEDGER
if "%CHOICE%"=="0" goto BACK

echo %RED%Invalid choice%RST%
timeout /t 2 >nul
goto :EOF

:CHECK_SERVICES
cls
echo %GRN%[INFO] Checking nexa.exe status...%RST%
echo.
tasklist | findstr /I "nexa.exe"
if errorlevel 1 (
    echo %RED%✗ nexa.exe is NOT running%RST%
    echo %GRA%Run: bin\start-all.bat%RST%
) else (
    echo %GRN%✓ nexa.exe is running%RST%
    echo %GRA%Checking ports...%RST%
    netstat -ano | findstr :7000
)
echo.
pause
goto :EOF

:KILL_SERVICES
cls
echo %YLW%[WARN] Terminating nexa.exe...%RST%
taskkill /F /IM nexa.exe >nul 2>&1
if errorlevel 1 (
    echo %YLW%⚠ nexa.exe was not running%RST%
) else (
    echo %GRN%✓ nexa.exe terminated%RST%
)
timeout /t 2 >nul
goto :EOF

:CLEAN_BUILD
cls
echo %YLW%[STEP 1] Stopping nexa.exe...%RST%
taskkill /F /IM nexa.exe >nul 2>&1
timeout /t 1 >nul

echo %YLW%[STEP 2] Cleaning build cache...%RST%
cd /d "%~dp0.."
go clean >nul 2>&1
del /F /Q "bin\nexa.exe" >nul 2>&1
echo %GRN%✓ Cache cleared%RST%

echo %YLW%[STEP 3] Rebuilding unified core...%RST%
echo %GRA%This may take 30-60 seconds...%RST%
echo.
go build -o "bin/nexa.exe" ".\cmd\nexa"
if !errorlevel! neq 0 (
    echo %RED%✗ Build FAILED - check Go installation%RST%
    timeout /t 3 >nul
    goto :EOF
)

if not exist "bin\nexa.exe" (
    echo %RED%✗ Build completed but nexa.exe not found%RST%
    timeout /t 3 >nul
    goto :EOF
)

echo.
echo %GRN%✓ Build successful!%RST%
echo %GRA%All 8 services compiled into bin\nexa.exe%RST%
echo %GRA%Run: bin\start-all.bat%RST%
timeout /t 2 >nul
goto :EOF

:CHECK_PORTS
cls
echo %CYN%Checking all service ports...%RST%
echo.
echo %GRA%Primary Services:%RST%
netstat -ano | findstr :7000 >nul && echo %GRN%✓ :7000 Dashboard%RST% || echo %GRA%○ :7000 Dashboard (free)%RST%
netstat -ano | findstr :8000 >nul && echo %GRN%✓ :8000 Gateway%RST% || echo %GRA%○ :8000 Gateway (free)%RST%
netstat -ano | findstr :8080 >nul && echo %GRN%✓ :8080 Admin%RST% || echo %GRA%○ :8080 Admin (free)%RST%
echo.
echo %GRA%Storage & Communication:%RST%
netstat -ano | findstr :8081 >nul && echo %GRN%✓ :8081 Storage%RST% || echo %GRA%○ :8081 Storage (free)%RST%
netstat -ano | findstr :8082 >nul && echo %GRN%✓ :8082 Chat%RST% || echo %GRA%○ :8082 Chat (free)%RST%
netstat -ano | findstr :3000 >nul && echo %GRN%✓ :3000 Web%RST% || echo %GRA%○ :3000 Web (free)%RST%
echo.
echo %GRA%Backend Services:%RST%
netstat -ano | findstr :1413 >nul && echo %GRN%✓ :1413 Core Server%RST% || echo %GRA%○ :1413 Core Server (free)%RST%
netstat -ano | findstr :1112 >nul && echo %GRN%✓ :1112 DNS Server%RST% || echo %GRA%○ :1112 DNS Server (free)%RST%
echo.
pause
goto :EOF

:VIEW_LOGS
cls
echo %CYN%Checking log files...%RST%
echo.
if exist "config\config.json" (
    echo %GRA%config.json found%RST%
)
if exist "config\users.json" (
    echo %GRA%users.json found%RST%
)
if exist "ledger.json" (
    echo %GRA%✓ ledger.json exists%RST%
    for %%A in (ledger.json) do echo   Size: %%~zA bytes
) else (
    echo %YLW%⚠ ledger.json not found (will be created on first run)%RST%
)
echo.
if exist "dns_records.json" (
    echo %GRA%✓ dns_records.json exists%RST%
)
echo.
echo %GRA%Try viewing live logs from the running system.%RST%
pause
goto :EOF

:RUN_DEBUG
cls
echo %CYN%Starting nexa.exe in DEBUG mode%RST%
echo %GRA%Press Ctrl+C to stop%RST%
echo %RED%═══════════════════════════════════════════════════════%RST%
echo.
cd /d "%~dp0\..\bin"
echo %GRA%Running: nexa.exe%RST%
echo.
nexa.exe
echo.
echo %RED%═══════════════════════════════════════════════════════%RST%
pause
goto :EOF

:CHECK_GO
cls
echo %CYN%Verifying Go installation...%RST%
echo.
go version
if errorlevel 1 (
    echo %RED%✗ Go is NOT installed or not in PATH%RST%
    echo %GRA%Download: https://golang.org/dl/%RST%
) else (
    echo %GRN%✓ Go detected%RST%
)
echo.
go env GOROOT
echo.
go env GOPATH
echo.
pause
goto :EOF

:CHECK_LEDGER
cls
echo %CYN%Checking ledger.json (blockchain)...%RST%
echo.
cd /d "%~dp0\.."
if not exist "ledger.json" (
    echo %YLW%⚠ ledger.json not found%RST%
    echo %GRA%It will be created automatically on first run.%RST%
) else (
    echo %GRN%✓ ledger.json found%RST%
    echo.
    echo %GRA%File size:%RST%
    for %%A in (ledger.json) do echo   %%~zA bytes
    echo.
    echo %GRA%Modified:%RST%
    for %%A in (ledger.json) do echo   %%~TA
)
echo.
pause
goto :EOF

:BACK
exit /b

```
