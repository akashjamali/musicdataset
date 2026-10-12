-- ======================================================================
-- HELLOSINDH MUSIC - MYSQL DATABASE INITIALIZATION
-- ======================================================================

CREATE DATABASE IF NOT EXISTS shazam CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE shazam;

-- 1. Songs table
CREATE TABLE IF NOT EXISTS songs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    artist VARCHAR(255) NOT NULL,
    yt_url VARCHAR(500) NOT NULL,
    spotify_url VARCHAR(500) DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_yt_url (yt_url(255)),
    INDEX idx_title (title(255)),
    INDEX idx_artist (artist)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. Fingerprints table
CREATE TABLE IF NOT EXISTS fingerprints (
    hash BIGINT NOT NULL,
    song_id INT NOT NULL,
    time_offset_ms INT NOT NULL,
    INDEX idx_hash (hash),
    INDEX idx_song_id (song_id),
    CONSTRAINT fk_fingerprints_song FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
