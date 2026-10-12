package main

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Danztee/shazam-build/internal/audio"
	"github.com/Danztee/shazam-build/internal/env"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type SongEntry struct {
	Link            string `json:"link"`
	VideoID         string `json:"videoId"`
	Title           string `json:"title"`
	Channel         string `json:"channel"`
	ChannelID       string `json:"channelId"`
	PublishedAt     string `json:"publishedAt"`
	DurationSeconds int    `json:"duration_seconds"`
	Views           int64  `json:"views"`
	MatchedQuery    string `json:"matched_query"`
}

type Dataset struct {
	Songs []SongEntry `json:"songs"`
}

func main() {
	jsonPath := flag.String("dataset", "sindhi_songs_dataset.json", "Path to dataset json")
	limitFlag := flag.Int("limit", 0, "Number of songs to ingest (0 = all)")
	workersFlag := flag.Int("workers", 8, "Number of parallel workers")
	flag.Parse()

	_ = godotenv.Load()

	dataBytes, err := os.ReadFile(*jsonPath)
	if err != nil {
		slog.Error("failed to read dataset file", "err", err)
		os.Exit(1)
	}

	var dataset Dataset
	if err := json.Unmarshal(dataBytes, &dataset); err != nil {
		slog.Error("failed to parse dataset JSON", "err", err)
		os.Exit(1)
	}

	targetLimit := *limitFlag
	if targetLimit <= 0 || targetLimit > len(dataset.Songs) {
		targetLimit = len(dataset.Songs)
	}

	slog.Info("Starting Fast Parallel Scraper",
		"total_in_dataset", len(dataset.Songs),
		"target_limit", targetLimit,
		"workers", *workersFlag,
	)

	// Connect to MySQL with local infile allowed
	dsn := env.GetString("DATABASE_URL", "root:@tcp(127.0.0.1:3306)/shazam?allowAllFiles=true&parseTime=true")
	if !strings.Contains(dsn, "allowAllFiles=true") {
		if strings.Contains(dsn, "?") {
			dsn += "&allowAllFiles=true"
		} else {
			dsn += "?allowAllFiles=true"
		}
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("failed to open MySQL", "err", err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(*workersFlag * 3)
	db.SetMaxIdleConns(*workersFlag * 2)

	if err := db.Ping(); err != nil {
		slog.Error("failed to connect to MySQL", "err", err)
		os.Exit(1)
	}

	audioSvc := audio.NewService(slog.Default())

	jobs := make(chan SongEntry, targetLimit)
	for i := 0; i < targetLimit; i++ {
		jobs <- dataset.Songs[i]
	}
	close(jobs)

	var (
		ingestedCount atomic.Int64
		skippedCount  atomic.Int64
		failedCount   atomic.Int64
		wg            sync.WaitGroup
		dbLock        sync.Mutex
	)

	for w := 1; w <= *workersFlag; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for s := range jobs {
				ytURL := s.Link
				if ytURL == "" && s.VideoID != "" {
					ytURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", s.VideoID)
				}

				// Check if already in MySQL
				var existingID int
				err := db.QueryRow("SELECT id FROM songs WHERE yt_url = ? OR yt_url LIKE ? LIMIT 1",
					ytURL, "%"+s.VideoID+"%").Scan(&existingID)
				if err == nil && existingID > 0 {
					skippedCount.Add(1)
					continue
				}

				cleanTitle, artistName := parseTitleAndArtist(s.Title, s.Channel, s.MatchedQuery)
				spotifyURL := fmt.Sprintf("https://open.spotify.com/search/%s", url.QueryEscape(artistName+" "+cleanTitle))

				// 1. Download WAV with strict 75s timeout to prevent hanging on stalled streams
				dlCtx, dlCancel := context.WithTimeout(context.Background(), 75*time.Second)
				wavPath, err := downloadAndConvertToWav(dlCtx, s.VideoID, ytURL)
				dlCancel()
				if err != nil {
					slog.Warn("download failed, skipping", "worker", workerID, "id", s.VideoID, "err", err)
					failedCount.Add(1)
					continue
				}

				// 2. Compute fingerprints
				fps, err := audioSvc.ProcessAudio(context.Background(), wavPath)
				_ = os.Remove(wavPath)
				if err != nil || len(fps) == 0 {
					slog.Warn("fingerprint failed, skipping", "worker", workerID, "id", s.VideoID, "err", err)
					failedCount.Add(1)
					continue
				}

				// 3. Deduplicate
				type fpKey struct {
					Hash   int64
					Offset int32
				}
				seen := make(map[fpKey]bool, len(fps))
				var deduped []audio.Fingerprint
				for _, fp := range fps {
					k := fpKey{Hash: fp.Hash, Offset: int32(fp.TimeOffsetMs)}
					if !seen[k] {
						seen[k] = true
						deduped = append(deduped, fp)
					}
				}

				// 4. Insert song record
				artistsJSON, _ := json.Marshal([]string{artistName})
				var songID int32

				dbLock.Lock()
				res, err := db.Exec(
					"INSERT INTO songs (title, artists, album, yt_url, spotify_url) VALUES (?, ?, ?, ?, ?)",
					cleanTitle, string(artistsJSON), "سنڌي ميوزڪ گڏهه", ytURL, spotifyURL,
				)
				if err != nil {
					dbLock.Unlock()
					slog.Error("failed to insert song", "worker", workerID, "err", err)
					failedCount.Add(1)
					continue
				}
				songID64, _ := res.LastInsertId()
				songID = int32(songID64)
				dbLock.Unlock()

				// 5. Bulk write fingerprints via TSV file and LOAD DATA INFILE
				tsvPath := filepath.Join(os.TempDir(), fmt.Sprintf("fp_batch_%s_%d.tsv", s.VideoID, songID))
				tsvFile, err := os.Create(tsvPath)
				if err != nil {
					slog.Error("failed to create temp TSV", "err", err)
					failedCount.Add(1)
					continue
				}

				buf := bufio.NewWriterSize(tsvFile, 64*1024)
				for _, fp := range deduped {
					fmt.Fprintf(buf, "%d\t%d\t%d\n", fp.Hash, songID, int32(fp.TimeOffsetMs))
				}
				buf.Flush()
				tsvFile.Close()

				cleanTsvPath := strings.ReplaceAll(tsvPath, "\\", "/")
				loadQuery := fmt.Sprintf("LOAD DATA LOCAL INFILE '%s' INTO TABLE fingerprints (hash, song_id, time_offset_ms)", cleanTsvPath)
				_, err = db.Exec(loadQuery)
				_ = os.Remove(tsvPath)

				if err != nil {
					slog.Error("LOAD DATA INFILE failed, fallback to chunk insert", "worker", workerID, "err", err)
					// Fallback chunk insert
					fallbackInsert(db, deduped, songID)
				}

				current := ingestedCount.Add(1)
				slog.Info("SUCCESS INGESTED",
					"worker", workerID,
					"song_id", songID,
					"title", cleanTitle,
					"artist", artistName,
					"fingerprints", len(deduped),
					"total_ingested", current,
					"target", targetLimit,
				)
			}
		}(w)
	}

	wg.Wait()
	fmt.Printf("\nDone! Ingested %d songs, skipped %d existing, failed %d.\n",
		ingestedCount.Load(), skippedCount.Load(), failedCount.Load())
}

func fallbackInsert(db *sql.DB, deduped []audio.Fingerprint, songID int32) {
	const batchSize = 3000
	for i := 0; i < len(deduped); i += batchSize {
		end := i + batchSize
		if end > len(deduped) {
			end = len(deduped)
		}
		chunk := deduped[i:end]

		var sb strings.Builder
		sb.WriteString("INSERT INTO fingerprints (hash, song_id, time_offset_ms) VALUES ")
		for j, fp := range chunk {
			if j > 0 {
				sb.WriteByte(',')
			}
			fmt.Fprintf(&sb, "(%d,%d,%d)", fp.Hash, songID, int32(fp.TimeOffsetMs))
		}
		db.Exec(sb.String())
	}
}

func parseTitleAndArtist(rawTitle, channel, query string) (cleanTitle, artist string) {
	t := html.UnescapeString(rawTitle)

	reNoise := regexp.MustCompile(`(?i)(official\s*(video|audio|music|hd|4k)?|full\s*(video|song|track)?|lyrical\s*video|prod\s*by.*|hd|hq|original|audio|video|mp3)`)
	t = reNoise.ReplaceAllString(t, "")
	reBrackets := regexp.MustCompile(`[\[\]\(\)\{\}\|\-_]`)
	t = reBrackets.ReplaceAllString(t, " ")
	t = strings.Join(strings.Fields(t), " ")

	artist = "سنڌي فنڪار"
	if strings.Contains(channel, "Official") || strings.Contains(channel, "Production") || strings.Contains(channel, "Enterprises") {
		if strings.Contains(t, "Mumtaz Molai") || strings.Contains(t, "ممتاز مولائي") {
			artist = "ممتاز مولائي"
		} else if strings.Contains(t, "Abida Parveen") || strings.Contains(t, "عابده پروين") {
			artist = "عابده پروين"
		} else if strings.Contains(t, "Sanam Marvi") || strings.Contains(t, "صنم ماروي") {
			artist = "صنم ماروي"
		} else if strings.Contains(t, "Sarmad Sindhi") || strings.Contains(t, "سرمد سنڌي") {
			artist = "سرمد سنڌي"
		} else if strings.Contains(t, "Manzoor Ali Khan") || strings.Contains(t, "منظور علي خان") {
			artist = "استاد منظور علي خان"
		} else if strings.Contains(t, "Muhammad Yousuf") || strings.Contains(t, "محمد يوسف") {
			artist = "استاد محمد يوسف"
		} else if strings.Contains(t, "Tufail") || strings.Contains(t, "طفيل") {
			artist = "طفيل سنجراڻي"
		} else if strings.Contains(t, "Jalal Chandio") || strings.Contains(t, "جلال چانڊيو") {
			artist = "جلال چانڊيو"
		} else if strings.Contains(t, "The Sketches") || strings.Contains(t, "Saif Samejo") {
			artist = "دي اسڪيچز"
		} else {
			artist = channel
		}
	} else if channel != "" {
		artist = channel
	}

	cleanTitle = t
	return
}

func downloadAndConvertToWav(ctx context.Context, videoID, ytURL string) (string, error) {
	tmpDir := os.TempDir()
	outWav := filepath.Join(tmpDir, fmt.Sprintf("dataset_%s_%d.wav", videoID, os.Getpid()))

	ytDlpPath := ".\\yt-dlp.exe"
	if _, err := os.Stat(ytDlpPath); err != nil {
		ytDlpPath = "yt-dlp"
	}

	cmd := exec.CommandContext(ctx, ytDlpPath,
		"--no-playlist",
		"--no-check-certificates",
		"--no-warnings",
		"--extractor-args", "youtube:player_skip=webpage,configs;player_client=android",
		"-f", "ba[ext=m4a]/ba[ext=webm]/ba/b",
		"--download-sections", "*00:00-03:30",
		"--socket-timeout", "15",
		"--retries", "2",
		"--fragment-retries", "2",
		"--extract-audio",
		"--audio-format", "wav",
		"--postprocessor-args", "ffmpeg:-ac 1 -ar 44100",
		"-o", outWav,
		ytURL,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w, out: %s", err, string(out))
	}
	return outWav, nil
}
