package store

import (
	"context"
	"time"
)

const DefaultLibraryID = "00000000-0000-0000-0000-000000000001"

type VideoStatus string

const (
	StatusPending     VideoStatus = "pending"
	StatusTranscoding VideoStatus = "transcoding"
	StatusCompleted   VideoStatus = "completed"
	StatusFailed      VideoStatus = "failed"
)

type Video struct {
	ID                    string      `json:"id"`
	LibraryID             string      `json:"library_id"`
	Title                 string      `json:"title"`
	OriginalExt           string      `json:"original_ext"`
	Status                VideoStatus `json:"status"`
	UploadSizeBytes       int64       `json:"upload_size_bytes"`
	TotalSizeBytes        int64       `json:"total_size_bytes"`
	Duration              int         `json:"duration"`
	TranscodingStartedAt  *time.Time  `json:"transcoding_started_at,omitempty"`
	TranscodingFinishedAt *time.Time  `json:"transcoding_finished_at,omitempty"`
	ErrorMessage          *string     `json:"error_message,omitempty"`
	CreatedAt             time.Time   `json:"created_at"`
	UpdatedAt             time.Time   `json:"updated_at"`
}

type Library struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type LibraryKey struct {
	ID             string    `json:"id"`
	LibraryID      string    `json:"library_id"`
	KeyName        string    `json:"key_name"`
	PlaybackSecret string    `json:"playback_secret"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
}

type Store interface {
	Create(ctx context.Context, v *Video) error
	Get(ctx context.Context, id string) (*Video, error)
	Update(ctx context.Context, v *Video) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Video, int, error)
	ListByLibrary(ctx context.Context, libraryID string, limit, offset int) ([]*Video, int, error)
	ListByStatus(ctx context.Context, statuses ...VideoStatus) ([]*Video, error)

	CreateLibrary(ctx context.Context, l *Library) error
	GetLibrary(ctx context.Context, id string) (*Library, error)
	ListLibraries(ctx context.Context) ([]*Library, error)
	UpdateLibrary(ctx context.Context, l *Library) error

	CreateLibraryKey(ctx context.Context, k *LibraryKey) error
	GetLibraryKey(ctx context.Context, id string) (*LibraryKey, error)
	ListLibraryKeys(ctx context.Context, libraryID string) ([]*LibraryKey, error)
	RevokeLibraryKey(ctx context.Context, id string) error
	RegenerateLibraryKey(ctx context.Context, id string, newSecret string) error
	DeleteLibraryKey(ctx context.Context, id string) error
}
