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

cls
echo.
echo %CYN%╔═══════════════════════════════════════════════════════════╗%RST%
echo %CYN%║        NEXA ULTIMATE v3.1 - System Launcher             ║%RST%
echo %CYN%╚═══════════════════════════════════════════════════════════╝%RST%
echo.
echo %GRA%Options:%RST%
echo.
echo   %CYN%1%RST% - Build & Launch (rebuild everything)
echo   %CYN%2%RST% - Launch (use existing binaries)
echo   %CYN%3%RST% - Troubleshooting
echo.
set /p CHOICE="%YLW%Select (1-3): %RST%"

if "%CHOICE%"=="1" goto BUILD_AND_RUN
if "%CHOICE%"=="2" goto RUN_ONLY
if "%CHOICE%"=="3" goto TROUBLESHOOT

echo %RED%Invalid choice%RST%
pause
exit /b 1

:BUILD_AND_RUN
cls
call BUILD.bat
exit /b

:RUN_ONLY
cls
cd /d "%~dp0"
call bin\start-all.bat
exit /b

:TROUBLESHOOT
cls
call TROUBLESHOOT.bat
exit /b
