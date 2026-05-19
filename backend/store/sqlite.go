package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

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
	query := `
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
	_, err := s.db.Exec(query)
	return err
}

func (s *SQLiteStore) Create(ctx context.Context, v *Video) error {
	v.CreatedAt = time.Now()
	v.UpdatedAt = v.CreatedAt
	query := `INSERT INTO videos (id, title, original_ext, status, upload_size_bytes, total_size_bytes, duration, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.ExecContext(ctx, query, v.ID, v.Title, v.OriginalExt, v.Status, v.UploadSizeBytes, v.TotalSizeBytes, v.Duration, v.CreatedAt, v.UpdatedAt)
	return err
}

func (s *SQLiteStore) Get(ctx context.Context, id string) (*Video, error) {
	v := &Video{}
	query := `SELECT id, title, original_ext, status, upload_size_bytes, total_size_bytes, duration, transcoding_started_at, transcoding_finished_at, error_message, created_at, updated_at FROM videos WHERE id = ?`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&v.ID, &v.Title, &v.OriginalExt, &v.Status, &v.UploadSizeBytes, &v.TotalSizeBytes, &v.Duration, &v.TranscodingStartedAt, &v.TranscodingFinishedAt, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt)
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

	rows, err := s.db.QueryContext(ctx, "SELECT id, title, original_ext, status, upload_size_bytes, total_size_bytes, duration, transcoding_started_at, transcoding_finished_at, error_message, created_at, updated_at FROM videos ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		v := &Video{}
		if err := rows.Scan(&v.ID, &v.Title, &v.OriginalExt, &v.Status, &v.UploadSizeBytes, &v.TotalSizeBytes, &v.Duration, &v.TranscodingStartedAt, &v.TranscodingFinishedAt, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt); err != nil {
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
		"SELECT id, title, original_ext, status, upload_size_bytes, total_size_bytes, duration, transcoding_started_at, transcoding_finished_at, error_message, created_at, updated_at FROM videos WHERE status IN (%s) ORDER BY created_at ASC",
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
		if err := rows.Scan(&v.ID, &v.Title, &v.OriginalExt, &v.Status, &v.UploadSizeBytes, &v.TotalSizeBytes, &v.Duration, &v.TranscodingStartedAt, &v.TranscodingFinishedAt, &v.ErrorMessage, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, rows.Err()
}
