# HelloSindh Music Scraper - Dubai High-Speed Deployment & Resume Guide

This package contains everything needed to run the HelloSindh Sindhi Music Scraper on another PC (e.g. in Dubai with high-speed internet) and transfer the completed database back.

---

## 📁 Package Contents

| File | Purpose |
|---|---|
| `dataset_scraper.exe` | Compiled high-speed Go scraper engine (8 parallel workers, automatic MySQL bulk insert via TSV). |
| `yt-dlp.exe` | Latest yt-dlp nightly binary (with anti-throttling and player_skip optimizations). |
| `ffmpeg.exe` & `ffprobe.exe` | Standalone audio converters (converts YouTube audio streams directly to 44.1kHz mono WAV). |
| `sindhi_songs_dataset.json` | Dataset containing 1,113 Sindhi songs with YouTube URLs and metadata. |
| `.env` | Environment configuration pointing to local MySQL (`DATABASE_URL=root:@tcp(127.0.0.1:3306)/shazam?allowAllFiles=true&parseTime=true`). |
| `schema.sql` | Database table structure for `songs` and `fingerprints`. |
| `init_database.bat` | One-click script to initialize the MySQL `shazam` database. |
| `scrape_dataset.bat` | One-click script to start scraping with 8 parallel workers. **Auto-skips any already ingested songs!** |
| `monitor_progress.bat` | Real-time command-line progress dashboard with live spinning heartbeat and song counter. |
| `live_progress.html` | Browser-based live dashboard (refreshing real-time stats). |
| `export_database.bat` | One-click script to dump the completed MySQL database into a `.sql` file to transfer back. |
| `import_database.bat` | One-click script to restore a `.sql` database dump. |

---

## 🚀 Setup on the Dubai PC (Step-by-Step)

### Step 1: Install & Start XAMPP (MySQL)
1. Download and install **XAMPP** for Windows (or standalone MySQL 8).
2. Open XAMPP Control Panel and click **Start** next to **MySQL**.
3. (Optional) If you brought over an existing SQL checkpoint file from the original PC:
   - Double click `init_database.bat` to create the `shazam` database.
   - Run `import_database.bat` and provide the `.sql` checkpoint filename so it doesn't re-download already processed songs!
   - If starting fresh, simply double-click `init_database.bat`.

### Step 2: Start Ingestion
Double click **`scrape_dataset.bat`**.
- It will launch `dataset_scraper.exe` with 8 parallel workers.
- It will automatically connect to MySQL and check `songs` table.
- Any song that is already present in MySQL will be **skipped in 0.001 seconds**.
- All new songs will be downloaded in 3.5-minute audio chunks, fingerprinted, and committed to MySQL instantly.
- **Drive space is fully protected:** Every downloaded `.wav` file is deleted immediately after fingerprinting.

### Step 3: Monitor In Real-Time
- Double click **`monitor_progress.bat`** to see a live terminal dashboard showing:
  - Progress bar (`[██████████░░░░░]`)
  - Total songs completed vs remaining (out of 1,113)
  - Last 5 songs added with titles and timestamps
  - Live heartbeat indicator

---

## ⏸️ How Resume Works (100% Safe)
- You can close the terminal or stop the scraper at **any time**.
- When you run `scrape_dataset.bat` again, the engine queries MySQL for existing `yt_url` entries and immediately resumes from where it left off.
- Zero duplicate songs or fingerprints will be generated.

---

## 📤 Transferring Database Back to Main PC

Once all 1,113 songs are ingested in Dubai:

1. On the Dubai PC, double-click **`export_database.bat`**.
   - It will generate a dump file named like `shazam_export_YYYYMMDD_HHMM.sql`.
2. Send that `.sql` file (or zip it) back to your main PC.
3. On your main PC:
   - Run `import_database.bat` and select the `.sql` file.
   - Your local MySQL will instantly have the full 1,113 Sindhi songs and all ~50 million fingerprints ready for instant recognition!
