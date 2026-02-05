@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title NEXA Troubleshooter v3.1

set "ESC= "
set "RED=%ESC%[91m"
set "GRN=%ESC%[92m"
set "YLW=%ESC%[93m"
set "CYN=%ESC%[96m"
set "GRA=%ESC%[90m"
set "RST=%ESC%[0m"

cls
echo.
echo %CYN%╔══════════════════════════════════════════════════════╗%RST%
echo %CYN%║            NEXA Troubleshooting Console              ║%RST%
echo %CYN%╚══════════════════════════════════════════════════════╝%RST%
echo.
echo %GRA%Select an option:%RST%
echo.
echo   %CYN%1%RST% %GRA%- Check if services are running%RST%
echo   %CYN%2%RST% %GRA%- Kill all NEXA processes%RST%
echo   %CYN%3%RST% %GRA%- Clean build (delete old .exe and rebuild)%RST%
echo   %CYN%4%RST% %GRA%- Test individual services%RST%
echo   %CYN%5%RST% %GRA%- Check port availability%RST%
echo   %CYN%6%RST% %GRA%- Run server in console (debug mode)%RST%
echo   %CYN%7%RST% %GRA%- Run gateway in console (debug mode)%RST%
echo.
set /p CHOICE="%YLW%Enter choice (1-7): %RST%"

if "%CHOICE%"=="1" goto CHECK_SERVICES
if "%CHOICE%"=="2" goto KILL_SERVICES
if "%CHOICE%"=="3" goto CLEAN_BUILD
if "%CHOICE%"=="4" goto TEST_SERVICES
if "%CHOICE%"=="5" goto CHECK_PORTS
if "%CHOICE%"=="6" goto RUN_SERVER
if "%CHOICE%"=="7" goto RUN_GATEWAY

echo %RED%Invalid choice%RST%
pause
exit /b 1

:CHECK_SERVICES
cls
echo %GRN%[INFO] Checking for running NEXA services...%RST%
echo.
tasklist | findstr /I "server.exe gateway.exe admin.exe web.exe dashboard.exe"
if errorlevel 1 (
    echo %RED%No NEXA services found running%RST%
) else (
    echo %GRN%Services detected above ✓%RST%
)
echo.
pause
goto :EOF

:KILL_SERVICES
cls
echo %YLW%[WARN] Terminating all NEXA services...%RST%
taskkill /F /IM server.exe >nul 2>&1
taskkill /F /IM gateway.exe >nul 2>&1
taskkill /F /IM admin.exe >nul 2>&1
taskkill /F /IM web.exe >nul 2>&1
taskkill /F /IM chat.exe >nul 2>&1
taskkill /F /IM dns.exe >nul 2>&1
taskkill /F /IM dashboard.exe >nul 2>&1
echo %GRN%[DONE] All services terminated%RST%
timeout /t 2 >nul
goto :EOF

:CLEAN_BUILD
cls
echo %YLW%[STEP 1] Killing existing services...%RST%
taskkill /F /IM server.exe >nul 2>&1
taskkill /F /IM gateway.exe >nul 2>&1
taskkill /F /IM admin.exe >nul 2>&1
taskkill /F /IM web.exe >nul 2>&1
taskkill /F /IM chat.exe >nul 2>&1
taskkill /F /IM dns.exe >nul 2>&1
taskkill /F /IM dashboard.exe >nul 2>&1
timeout /t 1 >nul

echo %YLW%[STEP 2] Cleaning build artifacts...%RST%
cd /d "%~dp0"
cd ..
go clean >nul 2>&1
del /F /Q bin\*.exe >nul 2>&1
echo %GRN%[DONE] Cleanup complete%RST%

echo %YLW%[STEP 3] Running fresh build...%RST%
echo %GRA%This may take 30-60 seconds...%RST%
echo.
go build -o bin\server.exe .\cmd\server
if errorlevel 1 (
    echo %RED%ERROR: Server build failed%RST%
    pause
    exit /b 1
)

go build -o bin\gateway.exe .\cmd\gateway
if errorlevel 1 (
    echo %RED%ERROR: Gateway build failed%RST%
    pause
    exit /b 1
)

echo.
echo %GRN%✓ Build successful!%RST%
echo %GRA%Run: .\bin\start-all.bat%RST%
pause
goto :EOF

:TEST_SERVICES
cls
echo %CYN%Testing individual services (each in new window)...%RST%
echo.
echo %YLW%[1] Starting Server (Ctrl+C to stop)...%RST%
start "Nexa Server (Debug)" cmd /k "cd bin && server.exe"
echo %GRA%Server started in new window%RST%
timeout /t 3 >nul

echo %YLW%[2] Starting Gateway (Ctrl+C to stop)...%RST%
start "Nexa Gateway (Debug)" cmd /k "cd bin && gateway.exe"
echo %GRA%Gateway started in new window%RST%
echo.
echo %GRN%[INFO] Both services should be running%RST%
echo %GRA%  - Server:  http://localhost:1413 (telnet test)%RST%
echo %GRA%  - Gateway: http://localhost:8000 (HTTP API)%RST%
echo.
pause
goto :EOF

:CHECK_PORTS
cls
echo %CYN%Checking port availability...%RST%
echo.
echo %GRA%═ TCP Ports ═%RST%
echo %GRA%Port 1413 (Server):%RST%
netstat -ano | findstr :1413
if errorlevel 1 echo   %GRN%✓ Available%RST%

echo %GRA%Port 8000 (Gateway):%RST%
netstat -ano | findstr :8000
if errorlevel 1 echo   %GRN%✓ Available%RST%

echo %GRA%Port 8080 (Admin):%RST%
netstat -ano | findstr :8080
if errorlevel 1 echo   %GRN%✓ Available%RST%

echo %GRA%Port 8081 (Web):%RST%
netstat -ano | findstr :8081
if errorlevel 1 echo   %GRN%✓ Available%RST%

echo %GRA%Port 7000 (Dashboard):%RST%
netstat -ano | findstr :7000
if errorlevel 1 echo   %GRN%✓ Available%RST%

echo.
echo %GRA%If a port is in use, you'll see output above.%RST%
echo %GRA%To kill process: taskkill /PID <PID> /F%RST%
echo.
pause
goto :EOF

:RUN_SERVER
cls
echo %CYN%Starting Server in DEBUG mode%RST%
echo %GRA%Press Ctrl+C to stop%RST%
echo %RED%═══════════════════════════════════════════════════════%RST%
echo.
cd /d "%~dp0\..\bin"
server.exe
pause
goto :EOF

:RUN_GATEWAY
cls
echo %CYN%Starting Gateway in DEBUG mode%RST%
echo %GRA%Press Ctrl+C to stop%RST%
echo %RED%═══════════════════════════════════════════════════════%RST%
echo.
cd /d "%~dp0\..\bin"
gateway.exe
pause
goto :EOF
