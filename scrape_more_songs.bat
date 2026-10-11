@echo off
chcp 65001 >nul
echo ======================================================================
echo           HELLOSINDH YOUTUBE SINDHI MUSIC SCRAPER & INGESTER
echo ======================================================================
echo.
echo Searching YouTube via API using sindhi_search_tags.json...
echo.
.\yt_scraper.exe -limit 5
echo.
echo Ingestion complete! Check view_database.bat or phpMyAdmin.
pause
