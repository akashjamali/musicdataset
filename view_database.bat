@echo off
chcp 65001 >nul

echo ======================================================================
echo                HELLOSINDH MUSIC DATABASE (MySQL / MariaDB)
echo ======================================================================
echo.
echo [1] SONGS IN DATABASE (with YouTube and Spotify URLs):
echo ----------------------------------------------------------------------
"C:\xampp\mysql\bin\mysql.exe" -u root --default-character-set=utf8mb4 shazam -e "SELECT id, title, artists, yt_url, spotify_url FROM songs ORDER BY id;"

echo.
echo [2] TOTAL AUDIO FINGERPRINTS STORED (fingerprints table):
echo ----------------------------------------------------------------------
"C:\xampp\mysql\bin\mysql.exe" -u root shazam -e "SELECT count(*) AS total_fingerprints FROM fingerprints;"

echo.
echo ======================================================================
echo Press any key to close this window...
pause >nul
