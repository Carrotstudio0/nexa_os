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
go build -o nexa.exe ./cmd/nexa/
if !errorLevel! neq 0 (
    echo  [ERROR] Main Go build failed!
    pause
    exit /b !errorLevel!
)

:: 3. Build Sub-Services
echo  [2/3] Compiling All Sub-Services (Gateway, Admin, DNS, Dashboard, etc.)...
go build -o bin/gateway.exe ./cmd/gateway/
go build -o bin/admin.exe ./cmd/admin/
go build -o bin/dns.exe ./cmd/dns/
go build -o bin/dashboard.exe ./cmd/dashboard/
go build -o bin/server.exe ./cmd/server/
go build -o bin/web.exe ./cmd/web/
go build -o bin/chat.exe ./cmd/chat/
go build -o bin/client.exe ./cmd/client/

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



