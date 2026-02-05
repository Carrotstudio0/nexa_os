@echo off
title Nexa Builder (Linux Cross-Compilation)
echo ===================================================
echo      NEXA SYSTEM - LINUX BUILDER
echo ===================================================
echo.

:: Create output directory
if not exist "bin\linux" mkdir "bin\linux"

:: Set Environment Variables for Linux Build
set GOOS=linux
set GOARCH=amd64
echo [SETUP] Target OS: Linux (amd64)

echo.
echo 1. Building Dashboard...
go build -o bin/linux/dashboard ./cmd/dashboard
if %errorlevel% neq 0 echo [X] Failed to build Dashboard & exit /b %errorlevel%

echo 2. Building Gateway...
go build -o bin/linux/gateway ./cmd/gateway
if %errorlevel% neq 0 echo [X] Failed to build Gateway & exit /b %errorlevel%

echo 3. Building File Manager (Web)...
go build -o bin/linux/web ./cmd/web
if %errorlevel% neq 0 echo [X] Failed to build Web & exit /b %errorlevel%

echo 3.5 Building Chat Service...
go build -o bin/linux/chat ./cmd/chat
if %errorlevel% neq 0 echo [X] Failed to build Chat & exit /b %errorlevel%

echo 4. Building Admin Panel...
go build -o bin/linux/admin ./cmd/admin
if %errorlevel% neq 0 echo [X] Failed to build Admin & exit /b %errorlevel%

echo 5. Building DNS Server...
go build -o bin/linux/dns ./cmd/dns
if %errorlevel% neq 0 echo [X] Failed to build DNS & exit /b %errorlevel%

echo 6. Building Core Server...
go build -o bin/linux/server ./cmd/server
if %errorlevel% neq 0 echo [X] Failed to build Server & exit /b %errorlevel%

echo.
echo ===================================================
echo [SUCCESS] Linux binaries created in: bin\linux\
echo ===================================================
echo.
pause
