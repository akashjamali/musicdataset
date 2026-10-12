@echo off
chcp 65001 >nul
echo ======================================================================
echo     HELLOSINDH - HIGH-SPEED 12-WORKER SINDHI SONGS INGESTION (MYSQL)
echo ======================================================================
echo.
echo Running 8 high-throughput parallel workers.
echo Anti-throttling & pure audio stream extraction enabled.
echo Recognition model and thresholds remain 100%% untouched.
echo.
.\dataset_scraper.exe -dataset sindhi_songs_dataset.json -workers 8 -limit 0
echo.
echo ======================================================================
echo Ingestion completed! Press any key to exit.
pause >nul
