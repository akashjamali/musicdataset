package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/Danztee/shazam-build/internal/audio"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	videoURL := "https://www.youtube.com/watch?v=ryzIL3GvNpk"
	testWav := "test_clip_11.wav"
	defer os.Remove(testWav)

	cmd := exec.Command(".\\yt-dlp.exe",
		"--extractor-args", "youtube:player_client=ios,android",
		"--extract-audio",
		"--audio-format", "wav",
		"--postprocessor-args", "ffmpeg:-ac 1 -ar 44100",
		"--download-sections", "*30-35",
		"-o", testWav,
		videoURL,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("yt-dlp failed: %v\n%s", err, string(out))
	}

	audioSvc := audio.NewService(nil)
	fps, err := audioSvc.ProcessAudio(context.Background(), testWav)
	if err != nil {
		log.Fatalf("audioSvc failed: %v", err)
	}
	fmt.Printf("Generated %d fingerprints from 5s clip\n", len(fps))

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/shazam")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	matchCount := 0
	offsets := make(map[int]int)
	for _, fp := range fps {
		rows, err := db.Query("SELECT time_offset_ms FROM fingerprints WHERE song_id = 11 AND hash = ?", fp.Hash)
		if err != nil {
			continue
		}
		for rows.Next() {
			var dbOffset int
			rows.Scan(&dbOffset)
			diff := dbOffset - fp.TimeOffsetMs
			bucket := diff / 100
			offsets[bucket]++
			matchCount++
		}
		rows.Close()
	}

	maxBucket := 0
	for _, c := range offsets {
		if c > maxBucket {
			maxBucket = c
		}
	}

	fmt.Printf("Total hash matches for Song 11: %d, Max diagonal alignment score: %d\n", matchCount, maxBucket)
}
