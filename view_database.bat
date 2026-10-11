@echo off
chcp 65001 >nul
set PGPASSWORD=postgres

echo ======================================================================
echo                      HELLOSINDH MUSIC DATABASE
echo ======================================================================
echo.
echo [1] SONGS IN DATABASE (songs table):
echo ----------------------------------------------------------------------
"C:\xampp\pgsql\bin\psql.exe" -U postgres -h localhost -p 5432 -d shazam -c "SELECT id, title, artists, album, created_at FROM songs ORDER BY id;"

echo.
echo [2] TOTAL AUDIO FINGERPRINTS STORED (fingerprints table):
echo ----------------------------------------------------------------------
"C:\xampp\pgsql\bin\psql.exe" -U postgres -h localhost -p 5432 -d shazam -c "SELECT count(*) AS total_fingerprints FROM fingerprints;"

echo.
echo ======================================================================
echo Press any key to close this window...
pause >nul
