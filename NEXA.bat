@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title NEXA ULTIMATE v3.1

:: Admin Check
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo.
    echo [!] Admin privileges required
    echo.
    pause
    exit /b 1
)

:: Colors
for /f %%a in ('echo prompt $E ^| cmd') do set "ESC=%%a"
set "BLU=%ESC%[94m"
set "GRA=%ESC%[90m"
set "GRN=%ESC%[92m"
set "YLW=%ESC%[93m"
set "RED=%ESC%[91m"
set "CYN=%ESC%[96m"
set "RST=%ESC%[0m"

:: CD to Script Directory
cd /d "%~dp0"

:: Check for CMD Arguments
if "%1"=="/run" goto RUN_ONLY
if "%1"=="/build" goto BUILD_AND_RUN

:: Environment Detection
set "IS_PROD=N"
if not exist "go.mod" (
    if exist "bin\nexa.exe" set "IS_PROD=Y"
    if exist "nexa.exe" set "IS_PROD=Y"
)

:MENU
cls
echo.
echo %CYN%╔═══════════════════════════════════════════════════════════╗%RST%
echo %CYN%║        NEXA ULTIMATE v3.1 - System Launcher             ║%RST%
echo %CYN%╚═══════════════════════════════════════════════════════════╝%RST%
echo.
if "!IS_PROD!"=="Y" (
    echo %GRN%[PRODUCTION MODE]%RST% %GRA%Detected installed environment%RST%
) else (
    echo %YLW%[DEVELOPMENT MODE]%RST% %GRA%Source code detected%RST%
)
echo.
echo %GRA%Options:%RST%
echo.

if "!IS_PROD!"=="Y" (
    echo   %CYN%1%RST% - Launch Nexa System
    echo   %CYN%2%RST% - Troubleshooting
    echo.
    set /p CHOICE="%YLW%Select (1-2): %RST%"
    if "!CHOICE!"=="1" goto RUN_ONLY
    if "!CHOICE!"=="2" goto TROUBLESHOOT
) else (
    echo   %CYN%1%RST% - Build ^& Launch (rebuild everything)
    echo   %CYN%2%RST% - Launch (use existing binaries)
    echo   %CYN%3%RST% - Troubleshooting
    echo.
    set /p CHOICE="%YLW%Select (1-3): %RST%"
    if "!CHOICE!"=="1" goto BUILD_AND_RUN
    if "!CHOICE!"=="2" goto RUN_ONLY
    if "!CHOICE!"=="3" goto TROUBLESHOOT
)


echo %RED%Invalid choice%RST%
timeout /t 2 >nul
goto MENU

:BUILD_AND_RUN
cls
if exist "scripts\build.bat" (
    call scripts\build.bat
) else (
    echo %RED%Error: scripts\build.bat not found.%RST%
    pause
)
exit /b

:RUN_ONLY
cls
if exist "bin\start-all.bat" (
    call bin\start-all.bat
) else if exist "start-all.bat" (
    call start-all.bat
) else (
    echo %RED%Error: start-all.bat not found.%RST%
    pause
)
exit /b

:TROUBLESHOOT
cls
if exist "scripts\troubleshoot.bat" (
    call scripts\troubleshoot.bat
) else (
    echo %RED%Error: scripts\troubleshoot.bat not found.%RST%
    pause
)
exit /b

