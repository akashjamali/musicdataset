# HelloSindh Music Scraper Checkpoint

**Timestamp:** 2026-10-12 20:17:00 PKT  
**Status:** **Stopped (Ready for instant resume)**  
**Target Dataset:** `sindhi_songs_dataset.json` (1,113 total songs)

---

## 1. Current State & Metrics

| Metric | Current Value |
|---|---|
| **Total Songs in MySQL** | **69 songs** |
| **Total Audio Fingerprints** | **~10,922,118** |
| **Database Engine** | MySQL 8 / MariaDB (XAMPP on `127.0.0.1:3306`) |
| **Database Name** | `shazam` |
| **Database Size** | **~1.01 GB** |
| **Recognition Model Threshold** | `bestScore < 50` (**100% original model preserved**) |
| **Drive C: Free Space** | **58.4 GB** (Zero leftover audio files on disk) |

---

## 2. Active Services

- **Backend Shazam Server:** `http://localhost:8080` (Running, connected to MySQL)
- **Frontend Web UI:** `http://localhost:5173` (Running, Vite dev server)
- **Database:** MySQL on `localhost:3306` (Database: `shazam`, User: `root`, Password: none)

---

## 3. How to Resume

Whenever you want to resume:
- Run `scrape_dataset.bat` or run:
```powershell
cd d:\hellosindh\hellosindhMusic
.\dataset_scraper.exe -dataset sindhi_songs_dataset.json -workers 8 -limit 0
```
It will automatically skip the 69 existing tracks and continue from song 70 onwards.
