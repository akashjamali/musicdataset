@echo off
chcp 65001 >nul
echo ======================================================================
echo       HELLOSINDH - INITIALIZE MYSQL DATABASE (SHAZAM)
echo ======================================================================
echo.

set MYSQL_BIN=C:\xampp\mysql\bin\mysql.exe
if not exist "%MYSQL_BIN%" (
    where mysql.exe >nul 2>&1
    if %errorlevel% equ 0 (
        set MYSQL_BIN=mysql.exe
    ) else (
        echo [ERROR] MySQL not found at C:\xampp\mysql\bin\mysql.exe or in PATH!
        echo Please ensure XAMPP MySQL is installed and running.
        pause
        exit /b 1
    )
)

echo Initializing database schema from schema.sql...
"%MYSQL_BIN%" -u root < schema.sql

if %errorlevel% equ 0 (
    echo.
    echo [SUCCESS] Database 'shazam' and tables created successfully!
) else (
    echo.
    echo [ERROR] Failed to initialize database. Is MySQL running in XAMPP?
)

echo.
pause
