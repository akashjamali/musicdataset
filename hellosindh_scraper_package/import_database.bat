@echo off
chcp 65001 >nul
echo ======================================================================
echo       HELLOSINDH - IMPORT SQL DUMP INTO MYSQL
echo ======================================================================
echo.

set MYSQL_BIN=C:\xampp\mysql\bin\mysql.exe
if not exist "%MYSQL_BIN%" (
    where mysql.exe >nul 2>&1
    if %errorlevel% equ 0 (
        set MYSQL_BIN=mysql.exe
    ) else (
        echo [ERROR] MySQL not found at C:\xampp\mysql\bin\mysql.exe or in PATH!
        pause
        exit /b 1
    )
)

if "%~1"=="" (
    echo Usage: Drag and drop your .sql file onto this batch file!
    echo Or specify the file name when prompted.
    echo.
    set /p SQL_FILE="Enter SQL filename to import (e.g. current_checkpoint.sql): "
) else (
    set SQL_FILE=%~1
)

if not exist "%SQL_FILE%" (
    echo [ERROR] File '%SQL_FILE%' not found!
    pause
    exit /b 1
)

echo Importing '%SQL_FILE%' into database 'shazam'...
"%MYSQL_BIN%" -u root --max_allowed_packet=512M shazam < "%SQL_FILE%"

if %errorlevel% equ 0 (
    echo.
    echo [SUCCESS] Import completed successfully!
) else (
    echo.
    echo [ERROR] Import failed!
)

echo.
pause
