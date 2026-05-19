package queue

import (
	"context"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/selvod/selvod/hooks"
	"github.com/selvod/selvod/store"
	"github.com/selvod/selvod/transcoder"
)

type WorkerPool struct {
	meta       store.Store
	transcoder transcoder.Transcoder
	hooks      *hooks.Registry
	storageDir string
	tasks      chan string
}

func NewWorkerPool(s store.Store, t transcoder.Transcoder, h *hooks.Registry, storageDir string, workers int) *WorkerPool {
	p := &WorkerPool{
		meta:       s,
		transcoder: t,
		hooks:      h,
		storageDir: storageDir,
		tasks:      make(chan string, 100),
	}

	for i := 0; i < workers; i++ {
		go p.worker()
	}
	return p
}

func (p *WorkerPool) Enqueue(id string) {
	p.tasks <- id
}

func (p *WorkerPool) worker() {
	for id := range p.tasks {
		p.process(id)
	}
}

func (p *WorkerPool) process(id string) {
	ctx := context.Background()
	v, err := p.meta.Get(ctx, id)
	if err != nil || v == nil {
		slog.Error("failed to fetch video for transcoding", "id", id, "error", err)
		return
	}

	now := time.Now()
	v.Status = store.StatusTranscoding
	v.TranscodingStartedAt = &now
	p.meta.Update(ctx, v)
	p.hooks.Dispatch(hooks.EventTranscodeStart, v)

	inputPath := filepath.Join(p.storageDir, "uploads", id+v.OriginalExt)
	outputDir := filepath.Join(p.storageDir, "hls", id)

	slog.Info("starting transcoding", "id", id, "input", inputPath)

	result, err := p.transcoder.Transcode(ctx, inputPath, outputDir)
	finishedAt := time.Now()
	v.TranscodingFinishedAt = &finishedAt

	if err != nil {
		slog.Error("transcoding failed", "id", id, "error", err)
		v.Status = store.StatusFailed
		errMsg := err.Error()
		if len(errMsg) > 200 {
			errMsg = errMsg[:200]
		}
		v.ErrorMessage = &errMsg
		p.meta.Update(ctx, v)
		p.hooks.Dispatch(hooks.EventTranscodeComplete, v)
		return
	}

	v.Status = store.StatusCompleted
	v.TotalSizeBytes = result.TotalSizeBytes
	if result.Duration > 0 {
		v.Duration = result.Duration
	}

	if err := p.meta.Update(ctx, v); err != nil {
		slog.Error("failed to update video after transcoding", "id", id, "error", err)
	}

	p.hooks.Dispatch(hooks.EventTranscodeComplete, v)
	slog.Info("transcoding completed", "id", id, "hls_size", v.TotalSizeBytes)
}

func (p *WorkerPool) Recover() {
	ctx := context.Background()
	videos, err := p.meta.ListByStatus(ctx, store.StatusPending, store.StatusTranscoding)
	if err != nil {
		slog.Error("failed to list videos for queue recovery", "error", err)
		return
	}

	for _, v := range videos {
		slog.Info("re-enqueuing recovered video task", "id", v.ID, "status", v.Status)
		p.Enqueue(v.ID)
	}
}
