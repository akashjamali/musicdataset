@echo off
chcp 65001 >nul
title HelloSindh Music - Live Ingestion Dashboard
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0monitor_progress.ps1"
if %ERRORLEVEL% neq 0 (
    echo.
    echo [ERROR] An error occurred while running the dashboard.
)
echo.
echo Press any key to exit...
pause >nul
