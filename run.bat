@echo off
echo Starting PostgreSQL from XAMPP...
start /b "" "C:\xampp\pgsql\bin\postgres.exe" -D "C:\xampp\pgsql\data"

timeout /t 2 /nobreak >nul

echo Starting Shazam Backend API (Port 8080)...
start /b "" "d:\hellosindh\hellosindhMusic\shazam_server.exe"

timeout /t 1 /nobreak >nul

echo Starting Shazam Frontend (Port 5173)...
cd /d "d:\hellosindh\hellosindhMusic\frontend"
npm run dev
