try { [Console]::OutputEncoding = [System.Text.Encoding]::UTF8 } catch {}
try { $Host.UI.RawUI.WindowTitle = "HelloSindh Music - Live Ingestion Dashboard" } catch {}

$mysqlPath = "C:\xampp\mysql\bin\mysql.exe"
$totalDataset = 1113

$cachedFpCount = "~13,000,000"
$cachedFpMB = "~1,100 MB"
$lastStatsCheck = [DateTime]::MinValue

$spinnerChars = @("|", "/", "-", "\")
$spinnerIdx = 0

function Get-ProgressReport {
    try { Clear-Host } catch { [Console]::Clear() }
    
    $global:spinnerIdx = ($global:spinnerIdx + 1) % $spinnerChars.Count
    $spinner = $spinnerChars[$global:spinnerIdx]

    Write-Host "==========================================================================" -ForegroundColor Cyan
    Write-Host "             HELLOSINDH MUSIC - LIVE INGESTION DASHBOARD" -ForegroundColor White
    Write-Host "==========================================================================" -ForegroundColor Cyan
    Write-Host ""

    # Check scraper process
    $scraperProc = Get-Process -Name "dataset_scraper" -ErrorAction SilentlyContinue
    $scraperStatus = if ($scraperProc) { "[ RUNNING - 8 WORKERS ACTIVE $spinner ]" } else { "[ PAUSED / IDLE ]" }
    $statusColor = if ($scraperProc) { "Green" } else { "Yellow" }

    Write-Host " HEARTBEAT STATUS:     " -NoNewline -ForegroundColor White
    Write-Host $scraperStatus -ForegroundColor $statusColor
    Write-Host " CURRENT TIME:         " -NoNewline -ForegroundColor White
    Write-Host (Get-Date -Format "hh:mm:ss tt") -ForegroundColor Gray
    Write-Host ""

    # Get fast song count
    $songsCountRaw = & $mysqlPath -u root shazam -s -N -e "SELECT count(*) FROM songs;" 2>$null
    $songsCount = if ($songsCountRaw) { [int]$songsCountRaw.Trim() } else { 0 }
    
    $remaining = [math]::Max(0, $totalDataset - $songsCount)
    $percent = [math]::Round(($songsCount / $totalDataset) * 100, 1)

    # Progress Bar
    $barLength = 32
    $filledLength = [math]::Min($barLength, [math]::Round(($percent / 100) * $barLength))
    $emptyLength = $barLength - $filledLength
    $bar = ("#" * $filledLength) + ("-" * $emptyLength)

    Write-Host " CATALOG PROGRESS:" -ForegroundColor Magenta
    Write-Host " ------------------------------------------------------------------------" -ForegroundColor DarkGray
    Write-Host "  Progress:  " -NoNewline
    Write-Host "[$bar] $percent%" -ForegroundColor Green
    Write-Host "  Completed: " -NoNewline
    Write-Host "$songsCount songs" -ForegroundColor Yellow -NoNewline
    Write-Host " of $totalDataset total" -ForegroundColor Gray
    Write-Host "  Remaining: " -NoNewline
    Write-Host "$remaining songs" -ForegroundColor Red
    Write-Host ""

    # Cached table stats (refresh every 30 seconds so it never delays console)
    if ((Get-Date).Subtract($global:lastStatsCheck).TotalSeconds -gt 30) {
        $stats = & $mysqlPath -u root shazam -s -N -e "SELECT table_rows, ROUND((data_length + index_length) / 1024 / 1024, 2) FROM information_schema.tables WHERE table_schema = 'shazam' AND table_name = 'fingerprints';" 2>$null
        if ($stats) {
            $parts = $stats -split "`t"
            if ($parts.Count -ge 2) {
                $global:cachedFpCount = ("{0:N0}" -f [int64]$parts[0])
                $global:cachedFpMB = "$($parts[1]) MB"
            }
        }
        $global:lastStatsCheck = Get-Date
    }

    Write-Host " DATABASE STORAGE:" -ForegroundColor Magenta
    Write-Host " ------------------------------------------------------------------------" -ForegroundColor DarkGray
    Write-Host "  Audio Fingerprints in MySQL: " -NoNewline
    Write-Host $global:cachedFpCount -ForegroundColor Cyan
    Write-Host "  Database Storage Size:       " -NoNewline
    Write-Host $global:cachedFpMB -ForegroundColor Cyan
    Write-Host ""

    # Fast latest 5 songs query
    Write-Host " RECENTLY INGESTED SONGS (LATEST 5):" -ForegroundColor Magenta
    Write-Host " ------------------------------------------------------------------------" -ForegroundColor DarkGray
    $recentSongs = & $mysqlPath -u root --default-character-set=utf8mb4 shazam -N -e "SELECT id, title, artists FROM songs ORDER BY id DESC LIMIT 5;" 2>$null
    if ($recentSongs) {
        $recentSongs -split "`n" | ForEach-Object {
            if ($_.Trim()) {
                $cols = $_ -split "`t"
                $id = $cols[0]
                $title = if ($cols.Count -gt 1) { $cols[1] } else { "" }
                $artist = if ($cols.Count -gt 2) { $cols[2].Replace('["','').Replace('"]','') } else { "" }
                Write-Host "  #$id " -ForegroundColor DarkCyan -NoNewline
                Write-Host "$title " -ForegroundColor White -NoNewline
                Write-Host "($artist)" -ForegroundColor Gray
            }
        }
    }
    Write-Host ""
    Write-Host "==========================================================================" -ForegroundColor Cyan
    Write-Host " NOTE: If terminal pauses, press [Enter] in this window (Windows QuickEdit)." -ForegroundColor DarkYellow
    Write-Host " Live heartbeat ticking every 2 seconds. Press [Ctrl+C] to close." -ForegroundColor DarkGray
}

while ($true) {
    try {
        Get-ProgressReport
    } catch {}
    Start-Sleep -Seconds 2
}
