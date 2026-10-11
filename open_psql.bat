@echo off
set PGPASSWORD=postgres
echo ==============================================
echo Connecting to PostgreSQL database 'shazam'...
echo ==============================================
"C:\xampp\pgsql\bin\psql.exe" -U postgres -h localhost -p 5432 -d shazam
