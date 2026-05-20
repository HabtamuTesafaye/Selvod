package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000;"); err != nil {
		return nil, fmt.Errorf("failed to set busy timeout: %w", err)
	}

	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return s, nil
}

func (s *SQLiteStore) migrate() error {
	// 1. Create libraries table
	librariesQuery := `
	CREATE TABLE IF NOT EXISTS libraries (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		created_at DATETIME NOT NULL
	);`
	if _, err := s.db.Exec(librariesQuery); err != nil {
		return fmt.Errorf("failed to create libraries table: %w", err)
	}

	// 2. Create library_keys table
	keysQuery := `
	CREATE TABLE IF NOT EXISTS library_keys (
		id TEXT PRIMARY KEY,
		library_id TEXT NOT NULL,
		key_name TEXT NOT NULL,
		playback_secret TEXT NOT NULL,
		is_active BOOLEAN NOT NULL DEFAULT 1,
		created_at DATETIME NOT NULL,
		FOREIGN KEY(library_id) REFERENCES libraries(id) ON DELETE CASCADE
	);`
	if _, err := s.db.Exec(keysQuery); err != nil {
		return fmt.Errorf("failed to create library_keys table: %w", err)
	}

	// 3. Create videos table if not exists
	videosQuery := `
	CREATE TABLE IF NOT EXISTS videos (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		original_ext TEXT NOT NULL,
		status TEXT NOT NULL,
		upload_size_bytes BIGINT NOT NULL,
		total_size_bytes BIGINT DEFAULT 0,
		duration INTEGER DEFAULT 0,
		transcoding_started_at DATETIME,
		transcoding_finished_at DATETIME,
		error_message TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`
	if _, err := s.db.Exec(videosQuery); err != nil {
		return fmt.Errorf("failed to create videos table: %w", err)
	}

	// 4. Check/Add library_id column to videos table
	var hasLibraryID bool
	rows, err := s.db.Query("PRAGMA table_info(videos)")
	if err != nil {
		return fmt.Errorf("failed to query videos table info: %w", err)
	}
	for rows.Next() {
		var cid int
		var name, typeStr string
		var notnull, pk int
		var dfltVal interface{}
		if err := rows.Scan(&cid, &name, &typeStr, &notnull, &dfltVal, &pk); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan table info row: %w", err)
		}
		if name == "library_id" {
			hasLibraryID = true
		}
	}
	rows.Close()

	if !hasLibraryID {
		if _, err := s.db.Exec("ALTER TABLE videos ADD COLUMN library_id TEXT;"); err != nil {
			return fmt.Errorf("failed to add library_id column to videos: %w", err)
		}
	}

	// 5. Seed default library with UUID
	if _, err := s.db.Exec(`INSERT OR IGNORE INTO libraries (id, name, created_at) VALUES (?, 'Default Library', ?)`, DefaultLibraryID, time.Now()); err != nil {
		return fmt.Errorf("failed to seed default library: %w", err)
	}

	// 6. Seed a default playback key if none exists
	var keyCount int
	s.db.QueryRow(`SELECT COUNT(*) FROM library_keys WHERE library_id = ?`, DefaultLibraryID).Scan(&keyCount)
	if keyCount == 0 {
		keyBytes := make([]byte, 16)
		rand.Read(keyBytes)
		secret := hex.EncodeToString(keyBytes)
		if _, err := s.db.Exec(`INSERT INTO library_keys (id, library_id, key_name, playback_secret, is_active, created_at) VALUES (?, ?, 'Default Playback Key', ?, 1, ?)`, uuid.New().String(), DefaultLibraryID, secret, time.Now()); err != nil {
			return fmt.Errorf("failed to seed default library key: %w", err)
		}
	}

	// 7. Backfill existing videos with default library UUID
	if _, err := s.db.Exec(`UPDATE videos SET library_id = ? WHERE library_id IS NULL OR library_id = ''`, DefaultLibraryID); err != nil {
		return fmt.Errorf("failed to backfill videos with default library id: %w", err)
	}

	return nil
}

func (s *SQLiteStore) Create(ctx context.Context, v *Video) error {
	v.CreatedAt = time.Now()
	v.UpdatedAt = v.CreatedAt
	if v.LibraryID == "" {
		v.LibraryID = DefaultLibraryID
	}
	query := `INSERT INTO videos (id, library_id, title, original_ext, status, upload_size_bytes, total_size_bytes, duration, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, v.ID, v.LibraryID, v.Title, v.OriginalExt, v.Status, v.UploadSizeBytes, v.TotalSizeBytes, v.Duration, v.CreatedAt, v.UpdatedAt)
	return err
}

func (s *SQLiteStore) Get(ctx context.Context, id string) (*Video, error) {
	v := &Video{}
	query := `SELECT id, library_id, title, original_ext, status, upload_size_bytes, total_size_bytes, duration, transcoding_started_at, transcoding_finished_at, error_message, created_at, updated_at FROM videos WHERE id = ?`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&v.ID, &v.LibraryID, &v.Title, &v.OriginalExt, &v.Status, &v.UploadSizeBytes, &v.TotalSizeBytes, &v.Duration, &v.TranscodingStartedAt, &v.TranscodingFinishedAt, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return v, err
}

func (s *SQLiteStore) Update(ctx context.Context, v *Video) error {
	v.UpdatedAt = time.Now()
	query := `UPDATE videos SET status = ?, total_size_bytes = ?, duration = ?, transcoding_started_at = ?, transcoding_finished_at = ?, error_message = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, v.Status, v.TotalSizeBytes, v.Duration, v.TranscodingStartedAt, v.TranscodingFinishedAt, v.ErrorMessage, v.UpdatedAt, v.ID)
	return err
}

func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM videos WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) List(ctx context.Context, limit, offset int) ([]*Video, int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM videos").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx, "SELECT id, library_id, title, original_ext, status, upload_size_bytes, total_size_bytes, duration, transcoding_started_at, transcoding_finished_at, error_message, created_at, updated_at FROM videos ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		v := &Video{}
		if err := rows.Scan(&v.ID, &v.LibraryID, &v.Title, &v.OriginalExt, &v.Status, &v.UploadSizeBytes, &v.TotalSizeBytes, &v.Duration, &v.TranscodingStartedAt, &v.TranscodingFinishedAt, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, err
		}
		videos = append(videos, v)
	}
	return videos, total, rows.Err()
}

func (s *SQLiteStore) ListByLibrary(ctx context.Context, libraryID string, limit, offset int) ([]*Video, int, error) {
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM videos WHERE library_id = ?", libraryID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := s.db.QueryContext(ctx, "SELECT id, library_id, title, original_ext, status, upload_size_bytes, total_size_bytes, duration, transcoding_started_at, transcoding_finished_at, error_message, created_at, updated_at FROM videos WHERE library_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", libraryID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		v := &Video{}
		if err := rows.Scan(&v.ID, &v.LibraryID, &v.Title, &v.OriginalExt, &v.Status, &v.UploadSizeBytes, &v.TotalSizeBytes, &v.Duration, &v.TranscodingStartedAt, &v.TranscodingFinishedAt, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, 0, err
		}
		videos = append(videos, v)
	}
	return videos, total, rows.Err()
}

func (s *SQLiteStore) ListByStatus(ctx context.Context, statuses ...VideoStatus) ([]*Video, error) {
	if len(statuses) == 0 {
		return []*Video{}, nil
	}

	placeholders := make([]string, len(statuses))
	args := make([]interface{}, len(statuses))
	for i, status := range statuses {
		placeholders[i] = "?"
		args[i] = status
	}

	query := fmt.Sprintf(
		"SELECT id, library_id, title, original_ext, status, upload_size_bytes, total_size_bytes, duration, transcoding_started_at, transcoding_finished_at, error_message, created_at, updated_at FROM videos WHERE status IN (%s) ORDER BY created_at ASC",
		strings.Join(placeholders, ","),
	)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		v := &Video{}
		if err := rows.Scan(&v.ID, &v.LibraryID, &v.Title, &v.OriginalExt, &v.Status, &v.UploadSizeBytes, &v.TotalSizeBytes, &v.Duration, &v.TranscodingStartedAt, &v.TranscodingFinishedAt, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, rows.Err()
}

func (s *SQLiteStore) CreateLibrary(ctx context.Context, l *Library) error {
	l.CreatedAt = time.Now()
	query := `INSERT INTO libraries (id, name, created_at) VALUES (?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, l.ID, l.Name, l.CreatedAt)
	return err
}

func (s *SQLiteStore) GetLibrary(ctx context.Context, id string) (*Library, error) {
	l := &Library{}
	query := `SELECT id, name, created_at FROM libraries WHERE id = ?`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&l.ID, &l.Name, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return l, err
}

func (s *SQLiteStore) ListLibraries(ctx context.Context) ([]*Library, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, created_at FROM libraries ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var libraries []*Library
	for rows.Next() {
		l := &Library{}
		if err := rows.Scan(&l.ID, &l.Name, &l.CreatedAt); err != nil {
			return nil, err
		}
		libraries = append(libraries, l)
	}
	return libraries, rows.Err()
}

func (s *SQLiteStore) CreateLibraryKey(ctx context.Context, k *LibraryKey) error {
	k.CreatedAt = time.Now()
	k.IsActive = true
	query := `INSERT INTO library_keys (id, library_id, key_name, playback_secret, is_active, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, k.ID, k.LibraryID, k.KeyName, k.PlaybackSecret, k.IsActive, k.CreatedAt)
	return err
}

func (s *SQLiteStore) GetLibraryKey(ctx context.Context, id string) (*LibraryKey, error) {
	k := &LibraryKey{}
	query := `SELECT id, library_id, key_name, playback_secret, is_active, created_at FROM library_keys WHERE id = ?`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&k.ID, &k.LibraryID, &k.KeyName, &k.PlaybackSecret, &k.IsActive, &k.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return k, err
}

func (s *SQLiteStore) ListLibraryKeys(ctx context.Context, libraryID string) ([]*LibraryKey, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, library_id, key_name, playback_secret, is_active, created_at FROM library_keys WHERE library_id = ? ORDER BY created_at DESC", libraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*LibraryKey
	for rows.Next() {
		k := &LibraryKey{}
		if err := rows.Scan(&k.ID, &k.LibraryID, &k.KeyName, &k.PlaybackSecret, &k.IsActive, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (s *SQLiteStore) RevokeLibraryKey(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE library_keys SET is_active = 0 WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) RegenerateLibraryKey(ctx context.Context, id string, newSecret string) error {
	_, err := s.db.ExecContext(ctx, "UPDATE library_keys SET playback_secret = ?, is_active = 1 WHERE id = ?", newSecret, id)
	return err
}

func (s *SQLiteStore) DeleteLibraryKey(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM library_keys WHERE id = ?", id)
	return err
}

func (s *SQLiteStore) UpdateLibrary(ctx context.Context, l *Library) error {
	query := `UPDATE libraries SET name = ? WHERE id = ?`
	_, err := s.db.ExecContext(ctx, query, l.Name, l.ID)
	return err
}
