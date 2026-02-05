@echo off
setlocal enabledelayedexpansion
title NEXA INSTALLER BUILD SYSTEM

:: الانتقال للفولدر الرئيسي للمشروع
cd /d "%~dp0.."

echo.
echo [1/3] Compiling NEXA Core Services...
echo --------------------------------------------------

:: بناء الخدمات الأساسية
set "SERVICES=nexa dns server gateway admin web dashboard chat"
for %%s in (%SERVICES%) do (
    echo   - Building %%s.exe...
    go build -o bin\%%s.exe .\cmd\%%s
    if !errorlevel! neq 0 (
        echo [!] Failed to build %%s. Exit.
        pause
        exit /b 1
    )
)

echo [+] Compilation complete.

echo.
echo [2/3] Searching for Inno Setup (ISCC.exe)...
echo --------------------------------------------------

set "ISCC_PATH="
:: محاولة إيجاد المسار يدوياً
if exist "C:\Program Files (x86)\Inno Setup 6\ISCC.exe" (
    set "ISCC_PATH=C:\Program Files (x86)\Inno Setup 6\ISCC.exe"
) else if exist "C:\Program Files\Inno Setup 6\ISCC.exe" (
    set "ISCC_PATH=C:\Program Files\Inno Setup 6\ISCC.exe"
) else (
    :: البحث في الـ PATH كخيار أخير
    for /f "tokens=*" %%i in ('where iscc.exe 2^>nul') do set "ISCC_PATH=%%i"
)

if "!ISCC_PATH!"=="" (
    echo [!] ERROR: Inno Setup (ISCC.exe) not found.
    echo [!] Please make sure Inno Setup is installed.
    pause
    exit /b 1
)

echo [+] Found Compiler at: !ISCC_PATH!

echo.
echo [3/3] Creating Windows Installer...
echo --------------------------------------------------
if not exist "installer_output" mkdir "installer_output"

:: تشغيل الكومبيلر
"!ISCC_PATH!" "installer\nexa.iss"

if %errorlevel% equ 0 (
    echo.
    echo ==================================================
    echo ✅ SUCCESS: Installer created!
    echo 📂 Path: \installer_output\Nexa_Setup_v3.1.exe
    echo ==================================================
    explorer "installer_output"
) else (
    echo.
    echo ❌ ERROR: Installer generation failed.
)

pause
