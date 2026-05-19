package store

import (
	"context"
	"time"
)

type VideoStatus string

const (
	StatusPending     VideoStatus = "pending"
	StatusTranscoding VideoStatus = "transcoding"
	StatusCompleted   VideoStatus = "completed"
	StatusFailed      VideoStatus = "failed"
)

type Video struct {
	ID                    string      `json:"id"`
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

type Store interface {
	Create(ctx context.Context, v *Video) error
	Get(ctx context.Context, id string) (*Video, error)
	Update(ctx context.Context, v *Video) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Video, int, error)
	ListByStatus(ctx context.Context, statuses ...VideoStatus) ([]*Video, error)
}
