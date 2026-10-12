package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Danztee/shazam-build/internal/audio"
	repo "github.com/Danztee/shazam-build/internal/database/queries"
	"github.com/Danztee/shazam-build/internal/env"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type TagsData struct {
	Generic struct {
		English []string `json:"english"`
		Sindhi  []string `json:"sindhi"`
	} `json:"generic"`
	Singers []struct {
		En        string   `json:"en"`
		Sd        string   `json:"sd"`
		AliasesEn []string `json:"aliases_en"`
	} `json:"singers"`
	Poets []struct {
		En        string   `json:"en"`
		Sd        string   `json:"sd"`
		AliasesEn []string `json:"aliases_en"`
	} `json:"poets"`
	QueryPatterns struct {
		English []string `json:"english"`
		Sindhi  []string `json:"sindhi"`
	} `json:"query_patterns"`
}

type YTSearchResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
		Snippet struct {
			Title        string `json:"title"`
			ChannelTitle string `json:"channelTitle"`
			Description  string `json:"description"`
		} `json:"snippet"`
	} `json:"items"`
}

func main() {
	limitFlag := flag.Int("limit", 4, "Number of songs to scrape and ingest")
	queryFlag := flag.String("query", "", "Specific search query (optional)")
	flag.Parse()

	_ = godotenv.Load()

	// 1. Read YouTube API key
	apiKeyBytes, err := os.ReadFile("ytapi")
	if err != nil {
		slog.Error("failed to read ytapi file", "err", err)
		os.Exit(1)
	}
	apiKey := strings.TrimSpace(string(apiKeyBytes))

	// 2. Read tags json
	tagsBytes, err := os.ReadFile("sindhi_search_tags.json")
	if err != nil {
		slog.Error("failed to read sindhi_search_tags.json", "err", err)
		os.Exit(1)
	}
	var tags TagsData
	if err := json.Unmarshal(tagsBytes, &tags); err != nil {
		slog.Error("failed to parse tags json", "err", err)
		os.Exit(1)
	}

	// 3. Connect to MySQL
	ctx := context.Background()
	dsn := env.GetString("DATABASE_URL", "root:@tcp(127.0.0.1:3306)/shazam?parseTime=true")
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("failed to open MySQL", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("failed to connect to MySQL", "err", err)
		os.Exit(1)
	}
	slog.Info("connected to MySQL database")

	audioSvc := audio.NewService(slog.Default())

	// Build search list
	type searchItem struct {
		Query      string
		SingerEn   string
		SingerSd   string
	}
	var searchQueue []searchItem

	if *queryFlag != "" {
		searchQueue = append(searchQueue, searchItem{
			Query:    *queryFlag,
			SingerEn: "Sindhi Artist",
			SingerSd: "سنڌي فنڪار",
		})
	} else {
		// Use singers from JSON file
		for _, s := range tags.Singers {
			query := fmt.Sprintf("%s جا بهترين گانا", s.Sd)
			searchQueue = append(searchQueue, searchItem{
				Query:    query,
				SingerEn: s.En,
				SingerSd: s.Sd,
			})
		}
	}

	ingestedCount := 0

	for _, item := range searchQueue {
		if ingestedCount >= *limitFlag {
			break
		}

		slog.Info("searching YouTube API", "query", item.Query)
		videos, err := searchYouTube(apiKey, item.Query, 3)
		if err != nil {
			slog.Warn("youtube search failed", "query", item.Query, "err", err)
			continue
		}

		for _, v := range videos {
			if ingestedCount >= *limitFlag {
				break
			}
			ytURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.ID)

			// Check if already in database
			var existingID int
			err := db.QueryRowContext(ctx, "SELECT id FROM songs WHERE yt_url = ? LIMIT 1", ytURL).Scan(&existingID)
			if err == nil {
				slog.Info("song already exists in database, skipping", "id", existingID, "url", ytURL)
				continue
			}

			// Clean and translate title to Sindhi
			cleanTitle := cleanSindhiTitle(v.Title, item.SingerSd, item.SingerEn)
			spotifyURL := fmt.Sprintf("https://open.spotify.com/search/%s", url.QueryEscape(item.SingerEn+" "+v.Title))

			slog.Info("processing new Sindhi song", "title", cleanTitle, "artist", item.SingerSd, "url", ytURL)

			// Download and convert audio to WAV
			wavPath, err := downloadAndConvertToWav(ctx, v.ID, ytURL)
			if err != nil {
				slog.Warn("failed to download audio", "id", v.ID, "err", err)
				continue
			}
			defer os.Remove(wavPath)

			// Generate fingerprints
			fingerprints, err := audioSvc.ProcessAudio(ctx, wavPath)
			if err != nil || len(fingerprints) == 0 {
				slog.Warn("fingerprinting failed", "id", v.ID, "err", err)
				os.Remove(wavPath)
				continue
			}

			// Insert into songs table
			artistsJSON, _ := json.Marshal([]string{item.SingerSd})
			res, err := db.ExecContext(ctx,
				"INSERT INTO songs (title, artists, album, yt_url, spotify_url) VALUES (?, ?, ?, ?, ?)",
				cleanTitle, string(artistsJSON), "سنڌي ميوزڪ البم", ytURL, spotifyURL,
			)
			if err != nil {
				slog.Error("failed to insert song", "err", err)
				os.Remove(wavPath)
				continue
			}

			songID64, _ := res.LastInsertId()
			songID := int32(songID64)

			// Deduplicate and save fingerprints
			type fpKey struct {
				Hash   int64
				Offset int32
			}
			seen := make(map[fpKey]bool)
			var deduped []audio.Fingerprint
			for _, fp := range fingerprints {
				k := fpKey{Hash: fp.Hash, Offset: int32(fp.TimeOffsetMs)}
				if !seen[k] {
					seen[k] = true
					deduped = append(deduped, fp)
				}
			}

			// Batch insert into MySQL
			const batchSize = 2000
			for i := 0; i < len(deduped); i += batchSize {
				end := i + batchSize
				if end > len(deduped) {
					end = len(deduped)
				}
				chunk := deduped[i:end]

				valStrings := make([]string, 0, len(chunk))
				valArgs := make([]interface{}, 0, len(chunk)*3)
				for _, fp := range chunk {
					valStrings = append(valStrings, "(?, ?, ?)")
					valArgs = append(valArgs, fp.Hash, songID, int32(fp.TimeOffsetMs))
				}

				stmt := fmt.Sprintf("INSERT INTO fingerprints (hash, song_id, time_offset_ms) VALUES %s", strings.Join(valStrings, ","))
				_, err := db.ExecContext(ctx, stmt, valArgs...)
				if err != nil {
					slog.Error("failed to insert fingerprints batch", "err", err)
					break
				}
			}

			os.Remove(wavPath)
			ingestedCount++
			slog.Info("SUCCESS: Ingested Sindhi song",
				"id", songID,
				"title", cleanTitle,
				"artist", item.SingerSd,
				"fingerprints", len(deduped),
				"total_ingested", ingestedCount,
			)
		}
	}

	fmt.Printf("\nDone! Successfully scraped and ingested %d new Sindhi songs into MySQL.\n", ingestedCount)
}

type videoInfo struct {
	ID    string
	Title string
}

func searchYouTube(apiKey, query string, maxResults int) ([]videoInfo, error) {
	reqURL := fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?part=snippet&type=video&videoCategoryId=10&maxResults=%d&q=%s&key=%s",
		maxResults, url.QueryEscape(query), apiKey)

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error %d: %s", resp.StatusCode, string(body))
	}

	var data YTSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var result []videoInfo
	for _, item := range data.Items {
		if item.ID.VideoID != "" {
			result = append(result, videoInfo{
				ID:    item.ID.VideoID,
				Title: item.Snippet.Title,
			})
		}
	}
	return result, nil
}

func cleanSindhiTitle(rawTitle, singerSd, singerEn string) string {
	t := rawTitle

	// Remove common YouTube noise words
	reNoise := regexp.MustCompile(`(?i)(official\s*(video|audio|music|hd|4k)?|full\s*(video|song|track)?|lyrical\s*video|prod\s*by.*|hd|hq|original|audio|video|mp3)`)
	t = reNoise.ReplaceAllString(t, "")

	// Remove brackets
	reBrackets := regexp.MustCompile(`[\[\]\(\)\{\}\|\-_]`)
	t = reBrackets.ReplaceAllString(t, " ")

	// Remove English singer name if present to keep it in Sindhi
	if singerEn != "" {
		reEn := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(singerEn))
		t = reEn.ReplaceAllString(t, "")
	}

	// Clean extra spaces
	t = strings.Join(strings.Fields(t), " ")

	if t == "" || len(t) < 3 {
		return fmt.Sprintf("%s جو گيت", singerSd)
	}

	// If title doesn't mention singer in Sindhi, prefix it
	if !strings.Contains(t, singerSd) {
		return fmt.Sprintf("%s - %s", singerSd, t)
	}

	return t
}

func downloadAndConvertToWav(ctx context.Context, videoID, ytURL string) (string, error) {
	tmpDir := os.TempDir()
	outWav := filepath.Join(tmpDir, fmt.Sprintf("scrape_%s.wav", videoID))

	// Find yt-dlp path
	ytDlpPath := ".\\yt-dlp.exe"
	if _, err := os.Stat(ytDlpPath); err != nil {
		ytDlpPath = "yt-dlp"
	}

	// Download audio to temp wav directly via yt-dlp + ffmpeg
	cmd := exec.CommandContext(ctx, ytDlpPath,
		"--extractor-args", "youtube:player_client=ios,android",
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

	if _, err := os.Stat(outWav); err != nil {
		return "", fmt.Errorf("output wav not found: %w", err)
	}

	return outWav, nil
}

// Suppress unused repo import warning
var _ = repo.CreateSongParams{}
