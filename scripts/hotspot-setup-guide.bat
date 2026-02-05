@echo off
setlocal enabledelayedexpansion
chcp 65001 >nul
title "NEXA WiFi Hotspot Setup Guide"

for /f %%a in ('echo prompt $E ^| cmd') do set "ESC=%%a"
set "BLU=%ESC%[94m"
set "GRA=%ESC%[90m"
set "GRN=%ESC%[92m"
set "YLW=%ESC%[93m"
set "RED=%ESC%[91m"
set "CYN=%ESC%[96m"
set "MAG=%ESC%[95m"
set "RST=%ESC%[0m"

cls
echo.
echo %CYN%╔════════════════════════════════════════════════════════╗%RST%
echo %CYN%║         NEXA WiFi Hotspot - Manual Setup Guide        ║%RST%
echo %CYN%╚════════════════════════════════════════════════════════╝%RST%
echo.
echo %GRA%إرشادات تفعيل الـ Hotspot يدويّاً من Windows%RST%
echo %GRA%══════════════════════════════════════════════════════════%RST%
echo.

echo %BLU%الخطوة 1: افتح إعدادات الشبكة%RST%
echo %GRA%─────────────────────────────────────%RST%
echo.
echo   اضغط على: %CYN%Windows Key + I%RST% (أو افتح Settings)
echo.
pause

cls

echo %BLU%الخطوة 2: انتقل إلى Mobile Hotspot%RST%
echo %GRA%─────────────────────────────────────%RST%
echo.
echo   1. اختر: %CYN%Network ^& Internet%RST%
echo   2. من اليسار: %CYN%Mobile Hotspot%RST%
echo.
pause

cls

echo %BLU%الخطوة 3: فعّل الـ Hotspot%RST%
echo %GRA%─────────────────────────────────────%RST%
echo.
echo   1. Toggle: %CYN%Share my internet connection%RST%
echo   2. اختر من القائمة: %CYN%Wi-Fi%RST% (ليس Bluetooth)
echo   3. اضغط: %GRN%Share%RST%
echo.
echo   ⏳ انتظر 2-3 ثوان حتى يتم التفعيل...
echo.
pause

cls

echo %BLU%الخطوة 4: تأكد من البيانات%RST%
echo %GRA%─────────────────────────────────────%RST%
echo.
echo   %GRA%يجب أن تشوف:%RST%
echo     • %CYN%Network name (SSID)%RST%: الاسم الموجود
echo     • %CYN%Password%RST%: كلمة المرور
echo.
echo   %YLW%اكتب هذه البيانات إذا احتجتها لاحقاً%RST%
echo.
pause

cls

echo %GRN%✓ تم تفعيل الـ Hotspot بنجاح!%RST%
echo %GRA%════════════════════════════════════════════════════════%RST%
echo.
echo %CYN%الآن قم بالخطوات التالية:%RST%
echo.
echo   1. عد إلى البرنامج (هذا الـ Window)
echo   2. اضغط %GRN%أي زر%RST% للمتابعة
echo.
pause

exit /b 0
