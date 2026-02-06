@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title "NEXA BUILD SYSTEM v3.1 - Unified Core"

:: Admin Check
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo [!] Admin privileges required
    pause
    exit /b 1
)

:: Colors
for /f %%a in ('echo prompt $E ^| cmd') do set "ESC=%%a"
set "BLU=%ESC%[94m"
set "GRA=%ESC%[90m"
set "GRN=%ESC%[92m"
set "RED=%ESC%[91m"
set "CYN=%ESC%[96m"
set "RST=%ESC%[0m"

cls
echo.
echo %CYN%NEXA ULTIMATE v3.1 - UNIFIED CORE BUILD%RST%
echo %GRA%════════════════════════════════════════════════════════%RST%
echo.

:: Navigate to project root
cd /d "%~dp0.."

:: Verify go.mod existence
if not exist "go.mod" (
    echo %RED%[!] ERROR: go.mod not found in %CD%%RST%
    echo %GRA%Ensure you are running the script from within the NEXA project structure.%RST%
    pause
    exit /b 1
)

:: Cleanup
echo %BLU%[CLEANUP]%RST% %GRA%Stopping any running services...%RST%
taskkill /F /IM nexa.exe >nul 2>&1
echo   %GRN%✓ Services terminated%RST%

if not exist "bin" mkdir "bin"
del /F /Q "bin\nexa.exe" >nul 2>&1
echo   %GRN%✓ Binary directory cleared%RST%

:: Verify
echo.
echo %BLU%[VERIFY]%RST% %GRA%Tidying Go modules...%RST%
go mod tidy >nul 2>&1
if !errorlevel! equ 0 (
    echo   %GRN%✓ Dependencies verified%RST%
) else (
    echo   %RED%✖ Module tidy failed%RST%
    pause
    exit /b 1
)

:: Build Unified Core
echo.
echo %BLU%[BUILD]%RST% %GRA%Compiling unified nexa.exe...%RST%
go build -o "bin/nexa.exe" ".\cmd\nexa"
if !errorlevel! neq 0 (
    echo   %RED%✖ Build FAILED%RST%
    echo   %GRA%Check your syntax and try again.%RST%
    pause
    exit /b 1
)
echo   %GRN%✓ nexa.exe compiled successfully%RST%

:: Verify Build
if not exist "bin\nexa.exe" (
    echo   %RED%✖ nexa.exe not found after build%RST%
    pause
    exit /b 1
)
echo   %GRN%✓ Build verification passed%RST%

:: Resources
echo.
echo %BLU%[RESOURCES]%RST% %GRA%Deploying configuration files...%RST%
if exist "config\config.json" copy /Y "config\config.json" "bin\" >nul
if exist "config\users.json" copy /Y "config\users.json" "bin\" >nul
if exist "users.json" copy /Y "users.json" "bin\" >nul
if exist "config.json" copy /Y "config.json" "bin\" >nul
echo   %GRN%✓ Config deployed%RST%

echo %BLU%[CERTS]%RST% %GRA%Deploying TLS certificates...%RST%
if not exist "bin\certs" mkdir "bin\certs"
if exist "certs" (
    copy /Y "certs\*.*" "bin\certs\" >nul
    echo   %GRN%✓ Certificates deployed%RST%
) else (
    echo   %YLW%⚠ Certificates not found (will use TCP fallback)%RST%
)

echo %BLU%[DEPLOY]%RST% %GRA%Finalizing deployment...%RST%
copy /Y "scripts\start-all.bat" "bin\" >nul
copy /Y "readme.md" "bin\" >nul
echo   %GRN%✓ Deployment files copied%RST%

echo.
echo %CYN%╔═══════════════════════════════════════════════════════════╗%RST%
echo %CYN%║         BUILD COMPLETED SUCCESSFULLY                     ║%RST%
echo %CYN%║  Unified Core: %GRN%bin\nexa.exe%CYN% (All 8 services included)    ║%RST%
echo %CYN%╚═══════════════════════════════════════════════════════════╝%RST%
echo.
echo %GRA%Services included in this build:%RST%
echo   %GRN%✓%RST% Dashboard (7000)    %GRN%✓%RST% Gateway (8000)      %GRN%✓%RST% Admin (8080)
echo   %GRN%✓%RST% Storage (8081)      %GRN%✓%RST% Chat (8082)        %GRN%✓%RST% DNS (1112)
echo   %GRN%✓%RST% Core Server (1413) %GRN%✓%RST% Web (3000)
echo.

set "START=N"
set /p START="%CYN%Launch system now? (Y/N): %RST%"
if /i "!START!"=="Y" (
    cls
    call bin\start-all.bat
) else (
    echo %GRA%[INFO] To launch, run: bin\start-all.bat%RST%
    echo %GRA%[INFO] Or use: NEXA.bat and select option 2%RST%
    pause
)

