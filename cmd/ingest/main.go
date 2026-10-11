package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Danztee/shazam-build/internal/audio"
	repo "github.com/Danztee/shazam-build/internal/database/queries"
	"github.com/Danztee/shazam-build/internal/env"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
)

func main() {
	wavPath := flag.String("wav", "", "Path to WAV file")
	title := flag.String("title", "", "Song title")
	artist := flag.String("artist", "Unknown", "Artist name")
	flag.Parse()

	if *wavPath == "" || *title == "" {
		fmt.Println("Usage: go run ./cmd/ingest -wav <path> -title <title> -artist <artist>")
		os.Exit(1)
	}

	_ = godotenv.Load()

	ctx := context.Background()
	dsn := env.GetString("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/shazam?sslmode=disable")

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		slog.Error("failed to connect to DB", "err", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	q := repo.New(conn)
	audioSvc := audio.NewService(slog.Default())

	// 1. Process audio fingerprints
	slog.Info("processing audio...", "file", *wavPath)
	fingerprints, err := audioSvc.ProcessAudio(ctx, *wavPath)
	if err != nil {
		slog.Error("failed to process audio", "err", err)
		os.Exit(1)
	}

	if len(fingerprints) == 0 {
		slog.Error("no fingerprints generated")
		os.Exit(1)
	}

	// 2. Insert song
	artistsJSON, _ := json.Marshal([]string{*artist})
	song, err := q.CreateSong(ctx, repo.CreateSongParams{
		Title:   *title,
		Artists: artistsJSON,
		Album:   pgtype.Text{String: "Sindhi Playlist", Valid: true},
	})
	if err != nil {
		slog.Error("failed to create song in DB", "err", err)
		os.Exit(1)
	}

	slog.Info("created song in database", "id", song.ID, "title", song.Title)

	// 3. Deduplicate fingerprints
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

	// 4. Batch insert via pgx CopyFrom
	slog.Info("inserting fingerprints...", "count", len(deduped))
	rowsCopied, err := conn.CopyFrom(
		ctx,
		pgx.Identifier{"fingerprints"},
		[]string{"hash", "song_id", "time_offset_ms"},
		pgx.CopyFromSlice(len(deduped), func(i int) ([]any, error) {
			return []any{
				deduped[i].Hash,
				song.ID,
				int32(deduped[i].TimeOffsetMs),
			}, nil
		}),
	)
	if err != nil {
		slog.Error("failed to batch insert fingerprints", "err", err)
		os.Exit(1)
	}

	slog.Info("SUCCESS: song ingested successfully!", "song_id", song.ID, "title", song.Title, "fingerprints_stored", rowsCopied)
}
