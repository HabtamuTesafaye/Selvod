package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/selvod/selvod/api"
	"github.com/selvod/selvod/config"
	"github.com/selvod/selvod/hooks"
	"github.com/selvod/selvod/queue"
	"github.com/selvod/selvod/signer"
	"github.com/selvod/selvod/store"
	"github.com/selvod/selvod/storage"
	"github.com/selvod/selvod/transcoder"
)

func main() {
	cfg := config.Load()

	st, err := storage.NewLocalStorage(cfg.StoragePath)
	if err != nil {
		slog.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}

	meta, err := store.NewSQLiteStore(cfg.DBPath)
	if err != nil {
		slog.Error("failed to initialize store", "error", err)
		os.Exit(1)
	}

	hr := hooks.NewRegistry()
	if cfg.WebhookURL != "" {
		hr.Add(hooks.NewWebhookHook(cfg.WebhookURL))
	}

	t := transcoder.NewFFmpegTranscoder()
	q := queue.NewWorkerPool(meta, t, hr, cfg.StoragePath, cfg.MaxWorkers)
	q.Recover()
	sig := signer.NewSecureSigner(cfg.StreamSecret, cfg.BaseURL)

	// Inject both the master APIKey and the independent PlaybackKey
	srv := api.NewServer(api.Config{
		Store:       meta,
		Storage:     st,
		Signer:      sig,
		Queue:       q,
		Hooks:       hr,
		Transcoder:  t,
		StorageDir:  cfg.StoragePath,
		APIKey:      cfg.APIKey,
		PlaybackKey: cfg.PlaybackKey,
	})

	go func() {
		if err := srv.Start(cfg.Port); err != nil && err != http.ErrServerClosed {
			slog.Error("server startup failed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("system shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		slog.Error("shutdown failed", "error", err)
	}
	if err := meta.Close(); err != nil {
		slog.Error("store close failed", "error", err)
	}
	slog.Info("shutdown complete")
}
