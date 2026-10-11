@echo off
chcp 65001 >nul
echo ==============================================
echo Connecting to MySQL database 'shazam'...
echo ==============================================
"C:\xampp\mysql\bin\mysql.exe" -u root --default-character-set=utf8mb4 shazam
