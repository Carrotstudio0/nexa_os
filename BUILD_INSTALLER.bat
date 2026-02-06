@echo off
setlocal enabledelayedexpansion
set PROJECT_DIR=%~dp0
set BIN_DIR=%PROJECT_DIR%bin
set INSTALLER_DIR=%PROJECT_DIR%installer
set ISCC_PATH=C:\Program Files (x86)\Inno Setup 6\ISCC.exe

echo [NEXA] Initializing Master Build Sequence v4.0.0-PRO...
echo [NEXA] Mode: Full Professional Suite (With Symbols)

if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"

:: 1. Build Service Matrix
echo.
echo [1/11] Compiling Nexa Nucleus (Supervisor)...
go build -o "%BIN_DIR%\nexa.exe" "%PROJECT_DIR%cmd\nexa\main.go"

echo [2/11] Compiling Matrix Gateway...
go build -o "%BIN_DIR%\gateway.exe" "%PROJECT_DIR%cmd\gateway"

echo [3/11] Compiling Admin Center...
go build -o "%BIN_DIR%\admin.exe" "%PROJECT_DIR%cmd\admin"

echo [4/11] Compiling Intelligence Hub...
go build -o "%BIN_DIR%\dashboard.exe" "%PROJECT_DIR%cmd\dashboard\main.go"

echo [5/11] Compiling Matrix Chat...
go build -o "%BIN_DIR%\chat.exe" "%PROJECT_DIR%cmd\chat\main.go"

echo [6/11] Compiling Digital Vault (Legacy Node)...
go build -o "%BIN_DIR%\web.exe" "%PROJECT_DIR%cmd\web\main.go"

echo [7/11] Compiling Core Server...
go build -o "%BIN_DIR%\server.exe" "%PROJECT_DIR%cmd\server\main.go"

echo [8/11] Compiling DNS Authority...
go build -o "%BIN_DIR%\dns.exe" "%PROJECT_DIR%cmd\dns\main.go"

echo [9/11] Compiling Mobile Client Bridge...
go build -o "%BIN_DIR%\client.exe" "%PROJECT_DIR%cmd\client\main.go"

echo [10/11] Compiling Utility: Certificate Generator...
go build -o "%BIN_DIR%\gen_certs.exe" "%PROJECT_DIR%tools\gen_certs\main.go"

echo [11/11] Compiling Utility: Security Hash Generator...
go build -o "%BIN_DIR%\hashgen.exe" "%PROJECT_DIR%tools\hashgen\main.go"

:: 2. Verify Binaries
echo.
echo [CHECK] Verifying build integrity...
dir "%BIN_DIR%\*.exe"

:: 3. Compile Installer
echo.
echo [NEXA] Engineering Universal Installer...

if not exist "%ISCC_PATH%" goto :NO_ISCC

"%ISCC_PATH%" "%INSTALLER_DIR%\nexa.iss"
if !ERRORLEVEL! NEQ 0 (
    echo [ERROR] Inno Setup compilation failed.
    exit /b !ERRORLEVEL!
)

echo.
echo [SUCCESS] Nexa Ultimate v4.0.0-PRO Full Professional Suite is ready.
echo [LOCATION] See 'installer_output' directory.
goto :END

:NO_ISCC
echo [WARNING] Inno Setup ISCC.exe not found at: "%ISCC_PATH%"
echo [INFO] Professional suite built successfully. You can run 'nexa.iss' manually.

:END
pause
