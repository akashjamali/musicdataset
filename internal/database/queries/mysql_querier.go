package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type MySQLQuerier struct {
	db *sql.DB
}

func NewMySQLQuerier(db *sql.DB) *MySQLQuerier {
	return &MySQLQuerier{db: db}
}

func (m *MySQLQuerier) CreateSong(ctx context.Context, arg CreateSongParams) (Song, error) {
	var album, imageUrl *string
	if arg.Album.Valid {
		album = &arg.Album.String
	}
	if arg.ImageUrl.Valid {
		imageUrl = &arg.ImageUrl.String
	}
	var durationMs *int32
	if arg.DurationMs.Valid {
		durationMs = &arg.DurationMs.Int32
	}

	res, err := m.db.ExecContext(ctx,
		"INSERT INTO songs (title, artists, album, duration_ms, image_url) VALUES (?, ?, ?, ?, ?)",
		arg.Title, string(arg.Artists), album, durationMs, imageUrl,
	)
	if err != nil {
		return Song{}, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return Song{}, err
	}

	return m.GetSong(ctx, int32(lastID))
}

func (m *MySQLQuerier) GetSong(ctx context.Context, id int32) (Song, error) {
	row := m.db.QueryRowContext(ctx,
		"SELECT id, title, artists, album, image_url, duration_ms, created_at FROM songs WHERE id = ? LIMIT 1",
		id,
	)
	return scanSong(row)
}

func (m *MySQLQuerier) GetSongByTitleAndArtists(ctx context.Context, arg GetSongByTitleAndArtistsParams) (Song, error) {
	row := m.db.QueryRowContext(ctx,
		"SELECT id, title, artists, album, image_url, duration_ms, created_at FROM songs WHERE title = ? AND JSON_CONTAINS(artists, ?) LIMIT 1",
		arg.Title, string(arg.Column2),
	)
	song, err := scanSong(row)
	if err != nil {
		// Fallback to title only
		row2 := m.db.QueryRowContext(ctx,
			"SELECT id, title, artists, album, image_url, duration_ms, created_at FROM songs WHERE title = ? LIMIT 1",
			arg.Title,
		)
		return scanSong(row2)
	}
	return song, nil
}

func (m *MySQLQuerier) ListSongs(ctx context.Context) ([]Song, error) {
	rows, err := m.db.QueryContext(ctx,
		"SELECT id, title, artists, album, image_url, duration_ms, created_at FROM songs ORDER BY id DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []Song
	for rows.Next() {
		s, err := scanSongFromRows(rows)
		if err != nil {
			return nil, err
		}
		songs = append(songs, s)
	}
	return songs, nil
}

func (m *MySQLQuerier) DeleteSong(ctx context.Context, id int32) error {
	_, err := m.db.ExecContext(ctx, "DELETE FROM songs WHERE id = ?", id)
	return err
}

func (m *MySQLQuerier) UpdateSong(ctx context.Context, arg UpdateSongParams) (Song, error) {
	var album, imageUrl *string
	if arg.Album.Valid {
		album = &arg.Album.String
	}
	if arg.ImageUrl.Valid {
		imageUrl = &arg.ImageUrl.String
	}
	var durationMs *int32
	if arg.DurationMs.Valid {
		durationMs = &arg.DurationMs.Int32
	}

	_, err := m.db.ExecContext(ctx,
		"UPDATE songs SET title = ?, artists = ?, album = ?, duration_ms = ?, image_url = ? WHERE id = ?",
		arg.Title, string(arg.Artists), album, durationMs, imageUrl, arg.ID,
	)
	if err != nil {
		return Song{}, err
	}
	return m.GetSong(ctx, arg.ID)
}

func (m *MySQLQuerier) CreateFingerprint(ctx context.Context, arg CreateFingerprintParams) (Fingerprint, error) {
	_, err := m.db.ExecContext(ctx,
		"INSERT INTO fingerprints (hash, song_id, time_offset_ms) VALUES (?, ?, ?)",
		arg.Hash, arg.SongID, arg.TimeOffsetMs,
	)
	if err != nil {
		return Fingerprint{}, err
	}
	return Fingerprint{
		Hash:         arg.Hash,
		SongID:       arg.SongID,
		TimeOffsetMs: arg.TimeOffsetMs,
	}, nil
}

func (m *MySQLQuerier) GetFingerprintsByHash(ctx context.Context, hash int64) ([]Fingerprint, error) {
	rows, err := m.db.QueryContext(ctx,
		"SELECT hash, song_id, time_offset_ms FROM fingerprints WHERE hash = ?",
		hash,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFingerprints(rows)
}

func (m *MySQLQuerier) GetFingerprintsByHashes(ctx context.Context, hashes []int64) ([]Fingerprint, error) {
	if len(hashes) == 0 {
		return nil, nil
	}

	const batchSize = 1000
	var allFingerprints []Fingerprint

	for i := 0; i < len(hashes); i += batchSize {
		end := i + batchSize
		if end > len(hashes) {
			end = len(hashes)
		}
		chunk := hashes[i:end]

		placeholders := strings.Repeat("?,", len(chunk))
		placeholders = placeholders[:len(placeholders)-1]

		args := make([]interface{}, len(chunk))
		for j, h := range chunk {
			args[j] = h
		}

		query := fmt.Sprintf("SELECT hash, song_id, time_offset_ms FROM fingerprints WHERE hash IN (%s)", placeholders)
		rows, err := m.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}

		fps, err := scanFingerprints(rows)
		rows.Close()
		if err != nil {
			return nil, err
		}
		allFingerprints = append(allFingerprints, fps...)
	}

	return allFingerprints, nil
}

func (m *MySQLQuerier) GetFingerprintsBySongID(ctx context.Context, songID int32) ([]Fingerprint, error) {
	rows, err := m.db.QueryContext(ctx,
		"SELECT hash, song_id, time_offset_ms FROM fingerprints WHERE song_id = ?",
		songID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanFingerprints(rows)
}

func (m *MySQLQuerier) DeleteFingerprintsBySongID(ctx context.Context, songID int32) error {
	_, err := m.db.ExecContext(ctx, "DELETE FROM fingerprints WHERE song_id = ?", songID)
	return err
}

func scanSong(row *sql.Row) (Song, error) {
	var s Song
	var artistsStr string
	var album, imageUrl sql.NullString
	var durationMs sql.NullInt32
	var createdAt time.Time

	err := row.Scan(&s.ID, &s.Title, &artistsStr, &album, &imageUrl, &durationMs, &createdAt)
	if err != nil {
		return Song{}, err
	}

	s.Artists = json.RawMessage(artistsStr)
	if album.Valid {
		s.Album = pgtype.Text{String: album.String, Valid: true}
	}
	if imageUrl.Valid {
		s.ImageUrl = pgtype.Text{String: imageUrl.String, Valid: true}
	}
	if durationMs.Valid {
		s.DurationMs = pgtype.Int4{Int32: durationMs.Int32, Valid: true}
	}
	s.CreatedAt = pgtype.Timestamp{Time: createdAt, Valid: true}

	return s, nil
}

func scanSongFromRows(rows *sql.Rows) (Song, error) {
	var s Song
	var artistsStr string
	var album, imageUrl sql.NullString
	var durationMs sql.NullInt32
	var createdAt time.Time

	err := rows.Scan(&s.ID, &s.Title, &artistsStr, &album, &imageUrl, &durationMs, &createdAt)
	if err != nil {
		return Song{}, err
	}

	s.Artists = json.RawMessage(artistsStr)
	if album.Valid {
		s.Album = pgtype.Text{String: album.String, Valid: true}
	}
	if imageUrl.Valid {
		s.ImageUrl = pgtype.Text{String: imageUrl.String, Valid: true}
	}
	if durationMs.Valid {
		s.DurationMs = pgtype.Int4{Int32: durationMs.Int32, Valid: true}
	}
	s.CreatedAt = pgtype.Timestamp{Time: createdAt, Valid: true}

	return s, nil
}

func scanFingerprints(rows *sql.Rows) ([]Fingerprint, error) {
	var items []Fingerprint
	for rows.Next() {
		var f Fingerprint
		if err := rows.Scan(&f.Hash, &f.SongID, &f.TimeOffsetMs); err != nil {
			return nil, err
		}
		items = append(items, f)
	}
	return items, rows.Err()
}
