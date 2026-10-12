@echo off
chcp 65001 >nul
echo ======================================================================
echo       HELLOSINDH - EXPORT MYSQL DATABASE TO SQL DUMP
echo ======================================================================
echo.

set DUMP_BIN=C:\xampp\mysql\bin\mysqldump.exe
if not exist "%DUMP_BIN%" (
    where mysqldump.exe >nul 2>&1
    if %errorlevel% equ 0 (
        set DUMP_BIN=mysqldump.exe
    ) else (
        echo [ERROR] mysqldump not found at C:\xampp\mysql\bin\mysqldump.exe or in PATH!
        pause
        exit /b 1
    )
)

set OUT_FILE=shazam_export_%date:~-4,4%%date:~-10,2%%date:~-7,2%_%time:~0,2%%time:~3,2%.sql
set OUT_FILE=%OUT_FILE: =0%

echo Dumping shazam database to %OUT_FILE% ...
"%DUMP_BIN%" -u root --single-transaction --quick --max_allowed_packet=512M shazam > "%OUT_FILE%"

if %errorlevel% equ 0 (
    echo.
    echo ======================================================================
    echo [SUCCESS] Database exported to: %OUT_FILE%
    echo You can now send this file back to your main PC!
    echo ======================================================================
) else (
    echo.
    echo [ERROR] Export failed!
)

echo.
pause
