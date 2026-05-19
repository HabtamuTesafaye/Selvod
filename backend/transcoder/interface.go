package transcoder

import "context"

type Result struct {
	TotalSizeBytes int64
	Duration       int
}

type Transcoder interface {
	Transcode(ctx context.Context, input, outputDir string) (*Result, error)
	IsVideo(ctx context.Context, input string) (bool, error)
}
