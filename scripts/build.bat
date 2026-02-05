@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title "NEXA BUILD SYSTEM v3.1"

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
echo %CYN%NEXA ULTIMATE v3.1 - BUILD SYSTEM%RST%
echo %GRA%════════════════════════════════════════════════════════%RST%
echo.

cd /d "%~dp0"
cd ..

:: Cleanup
echo %BLU%[CLEANUP]%RST% %GRA%Stopping services and clearing bin...%RST%
taskkill /F /IM nexa.exe >nul 2>&1
taskkill /F /IM dns.exe >nul 2>&1
taskkill /F /IM server.exe >nul 2>&1
taskkill /F /IM gateway.exe >nul 2>&1
taskkill /F /IM admin.exe >nul 2>&1
taskkill /F /IM web.exe >nul 2>&1
taskkill /F /IM dashboard.exe >nul 2>&1

if not exist bin mkdir bin
del /F /Q bin\*.exe >nul 2>&1

:: Verify
echo %BLU%[VERIFY]%RST% %GRA%Tidying modules...%RST%
go mod tidy >nul 2>&1

:: Build
echo %BLU%[BUILD]%RST% %GRA%Compiling all services...%RST%

set "SERVICES=nexa dns server gateway admin web dashboard chat"
for %%s in (%SERVICES%) do (
    echo   Compiling %%s.exe...
    go build -o bin\%%s.exe .\cmd\%%s
    if !errorlevel! neq 0 (
        echo   %RED%✖ Failed to build %%s%RST%
        pause
        exit /b 1
    )
)

:: Resources
echo.
echo %BLU%[RESOURCES]%RST% %GRA%Copying database and secure assets...%RST%
if exist "users.json" copy /Y "users.json" "bin\" >nul
if exist "config.json" copy /Y "config.json" "bin\" >nul
if not exist "bin\certs" mkdir "bin\certs"
if exist "certs" copy /Y "certs\*.*" "bin\certs\" >nul
copy /Y "scripts\start-all.bat" "bin\" >nul

echo   %GRN%✓ Assets deployed to \bin%RST%

echo.
echo %GRN%✓ ALL SYSTEMS BUILT SUCCESSFULLY%RST%
echo.

set "START=N"
set /p START="%CYN%Initialize launch sequence? (Y/N): %RST%"
if /i "!START!"=="Y" (
    cls
    call bin\start-all.bat
) else (
    echo %GRA%[INFO] Deployment ready in \bin. Run bin\start-all.bat to launch.%RST%
    pause
)
