@echo off
setlocal enabledelayedexpansion

:: NEXA INSTALLER COMPILER v4.0.0-PRO
:: COMPREHENSIVE SUITE BUILDER
title NEXA COMPREHENSIVE BUILDER

echo.
echo  [NEXA] Building Unified Pro Suite and All Services...
echo  --------------------------------------------------

:: 1. Ensure Bin directory exists
if not exist "bin" mkdir bin

:: 2. Build the Main Engine
echo  [1/3] Compiling Main Go Engine...
go build -trimpath -o nexa.exe ./cmd/nexa/
if !errorLevel! neq 0 (
    echo  [ERROR] Main Go build failed!
    pause
    exit /b !errorLevel!
)

:: 3. Build Sub-Services (Modular Components)
echo  [2/3] Compiling Modular Matrix Services...
echo      - Gateway Matrix...
go build -trimpath -o bin/nexa_gateway.exe ./cmd/gateway/
echo      - Admin Command Center...
go build -trimpath -o bin/nexa_admin.exe ./cmd/admin/
echo      - DNS Authority Node...
go build -trimpath -o bin/nexa_dns.exe ./cmd/dns/
echo      - Intelligence Dashboard...
go build -trimpath -o bin/nexa_dashboard.exe ./cmd/dashboard/
echo      - Core Server Engine...
go build -trimpath -o bin/nexa_core_server.exe ./cmd/server/
echo      - Universal Web Node...
go build -trimpath -o bin/nexa_web.exe ./cmd/web/
echo      - Quantum Chat Module...
go build -trimpath -o bin/nexa_chat.exe ./cmd/chat/
echo      - Terminal Client...
go build -trimpath -o bin/nexa_client.exe ./cmd/client/
echo.

:: 4. Locate and Run Inno Setup Compiler
echo  [3/3] Compiling Setup Executable...

set "ISCC_PATH="
if exist "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" set "ISCC_PATH=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
if defined ISCC_PATH goto :FoundISCC
if exist "C:\Program Files\Inno Setup 6\ISCC.exe" set "ISCC_PATH=C:\Program Files\Inno Setup 6\ISCC.exe"
if defined ISCC_PATH goto :FoundISCC
for /f "delims=" %%i in ('where iscc 2^>nul') do set "ISCC_PATH=%%i"
if defined ISCC_PATH goto :FoundISCC

:NoISCC
echo.
echo  [NOTICE] Inno Setup compiler (ISCC.exe) not found. 
echo  Please install Inno Setup 6 to generate the final .EXE
echo  You can download it from: https://jrsoftware.org/isdl.php
goto :EndProcess

:FoundISCC
echo  [INFO] Using Inno Setup from: "!ISCC_PATH!"
"!ISCC_PATH!" installer\nexa.iss
if !errorLevel! EQU 0 (
    echo.
    echo  [SUCCESS] Professional Installer is ready in 'installer_output' folder!
) else (
    echo  [ERROR] Inno Setup compilation failed with code !errorLevel!.
)

:EndProcess
echo.
echo  Process Complete.
pause



