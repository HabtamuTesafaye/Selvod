package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/selvod/selvod/hooks"
	"github.com/selvod/selvod/queue"
	"github.com/selvod/selvod/signer"
	"github.com/selvod/selvod/store"
	"github.com/selvod/selvod/storage"
	"github.com/selvod/selvod/transcoder"
)

type VideoHandler struct {
	Store        store.Store
	Storage      storage.Provider
	Signer       *signer.SecureSigner
	Queue        *queue.WorkerPool
	Hooks        *hooks.Registry
	Transcoder   transcoder.Transcoder
	StorageDir   string
	Cache        *LibraryCache
	GlobalSecret string
	configMutex  sync.Mutex
}

func (h *VideoHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 500<<20)
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		http.Error(w, "file too large or malformed", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file part in request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	id := uuid.New().String()
	ext := filepath.Ext(header.Filename)

	allowedExts := map[string]bool{
		".mp4": true, ".mov": true, ".mkv": true, ".webm": true,
		".mp3": true, ".aac": true, ".wav": true, ".ogg": true, ".flac": true,
	}
	if !allowedExts[ext] {
		http.Error(w, "file type not allowed", http.StatusUnsupportedMediaType)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		title = header.Filename
	}

	libraryID := r.FormValue("library_id")
	if libraryID == "" {
		libraryID = store.DefaultLibraryID
	}

	if ls, ok := h.Storage.(interface{ AvailableBytes() (int64, error) }); ok {
		if avail, err := ls.AvailableBytes(); err == nil && avail < 1<<30 {
			http.Error(w, "Storage critical: Upload rejected", http.StatusInsufficientStorage)
			return
		}
	}

	uploadPath := filepath.Join("uploads", id+ext)
	if err := h.Storage.Save(r.Context(), uploadPath, file); err != nil {
		http.Error(w, "failed to write file to storage", http.StatusInternalServerError)
		return
	}

	fullPath := filepath.Join(h.StorageDir, uploadPath)
	ok, probeErr := h.Transcoder.IsVideo(r.Context(), fullPath)
	if !ok {
		if err := h.Storage.Delete(r.Context(), uploadPath); err != nil {
			slog.Warn("failed to clean up rejected upload", "path", uploadPath, "error", err)
		}
		if probeErr != nil {
			slog.Error("ffprobe probe failed", "path", fullPath, "error", probeErr)
			http.Error(w, "failed to verify media file", http.StatusInternalServerError)
			return
		}
		http.Error(w, "Security rejection: Invalid or malicious media file", http.StatusUnsupportedMediaType)
		return
	}

	v := &store.Video{
		ID:              id,
		LibraryID:       libraryID,
		Title:           title,
		OriginalExt:     ext,
		Status:          store.StatusPending,
		UploadSizeBytes: header.Size,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.Store.Create(r.Context(), v); err != nil {
		http.Error(w, "failed to save metadata record", http.StatusInternalServerError)
		return
	}

	h.Queue.Enqueue(id)
	h.Hooks.Dispatch(hooks.EventUpload, v)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(v)
}

func (h *VideoHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	libraryID := r.URL.Query().Get("library_id")

	var videos []*store.Video
	var total int
	var err error

	if libraryID != "" {
		videos, total, err = h.Store.ListByLibrary(r.Context(), libraryID, 50, 0)
	} else {
		videos, total, err = h.Store.List(r.Context(), 50, 0)
	}

	if err != nil {
		http.Error(w, "failed to fetch video list", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"videos": videos,
		"total":  total,
	})
}

func (h *VideoHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := h.Store.Get(r.Context(), id)
	if err != nil || v == nil {
		http.Error(w, "video not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func (h *VideoHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := h.Store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "internal database error", http.StatusInternalServerError)
		return
	}
	if v == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.Storage.Delete(r.Context(), filepath.Join("uploads", id+v.OriginalExt)); err != nil {
		slog.Warn("failed to delete upload file", "id", id, "error", err)
	}
	if err := h.Storage.Delete(r.Context(), filepath.Join("libraries", v.LibraryID, "videos", id, "hls")); err != nil {
		slog.Warn("failed to delete HLS directory", "id", id, "error", err)
	}

	if err := h.Store.Delete(r.Context(), id); err != nil {
		http.Error(w, "failed to remove metadata record", http.StatusInternalServerError)
		return
	}
	h.Hooks.Dispatch(hooks.EventDelete, v)
	w.WriteHeader(http.StatusNoContent)
}

func (h *VideoHandler) playbackSecret(ctx context.Context, libraryID string) (string, error) {
	secret, ok := h.Cache.Get(libraryID)
	if ok {
		return secret, nil
	}
	keys, err := h.Store.ListLibraryKeys(ctx, libraryID)
	if err != nil {
		return "", fmt.Errorf("failed to list library keys: %w", err)
	}
	for _, k := range keys {
		if k.IsActive {
			h.Cache.Set(libraryID, k.PlaybackSecret, 5*time.Minute)
			return k.PlaybackSecret, nil
		}
	}
	return "", fmt.Errorf("no active playback keys for library %s", libraryID)
}

func (h *VideoHandler) HandleSign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := h.Store.Get(r.Context(), id)
	if err != nil || v == nil {
		http.Error(w, "video not found", http.StatusNotFound)
		return
	}

	if v.Status != store.StatusCompleted {
		http.Error(w, "video is still processing", http.StatusLocked)
		return
	}

	secret, err := h.playbackSecret(r.Context(), v.LibraryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	signed, err := h.Signer.Sign(v.LibraryID, id, secret, 2*time.Hour)
	if err != nil {
		http.Error(w, "failed to generate signed URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(signed)
}

func (h *VideoHandler) HandleEmbed(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := h.Store.Get(r.Context(), id)
	if err != nil || v == nil {
		http.Error(w, "video not found", http.StatusNotFound)
		return
	}

	if v.Status != store.StatusCompleted {
		http.Error(w, "video is still processing", http.StatusLocked)
		return
	}

	secret, err := h.playbackSecret(r.Context(), v.LibraryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// For embeds, we provide a longer duration signed URL (e.g. 24 hours)
	signed, err := h.Signer.Sign(v.LibraryID, id, secret, 24*time.Hour)
	if err != nil {
		http.Error(w, "failed to generate embed URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"url": signed.URL,
	})
}

func (h *VideoHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status": "ok",
		"time":   time.Now(),
	}
	if ls, ok := h.Storage.(interface{ AvailableBytes() (int64, error) }); ok {
		if available, err := ls.AvailableBytes(); err == nil {
			resp["storage_available_bytes"] = available
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *VideoHandler) HandleUpdateVideo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	v, err := h.Store.Get(r.Context(), id)
	if err != nil || v == nil {
		http.Error(w, "video not found", http.StatusNotFound)
		return
	}

	var input struct {
		Title     string `json:"title"`
		LibraryID string `json:"library_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	if input.Title != "" {
		v.Title = input.Title
	}

	if input.LibraryID != "" && input.LibraryID != v.LibraryID {
		// Verify new library exists
		lib, err := h.Store.GetLibrary(r.Context(), input.LibraryID)
		if err != nil || lib == nil {
			http.Error(w, "target library not found", http.StatusBadRequest)
			return
		}

		// Move HLS folder on disk if it exists
		oldPath := filepath.Join(h.StorageDir, "libraries", v.LibraryID, "videos", id)
		newParentPath := filepath.Join(h.StorageDir, "libraries", input.LibraryID, "videos")
		newPath := filepath.Join(newParentPath, id)

		// Ensure target directory's parent exists
		if err := os.MkdirAll(newParentPath, 0755); err != nil {
			http.Error(w, "failed to create target directory", http.StatusInternalServerError)
			return
		}

		// Rename directory if old exists
		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				http.Error(w, "failed to move video files on storage", http.StatusInternalServerError)
				return
			}
		}

		v.LibraryID = input.LibraryID
	}

	v.UpdatedAt = time.Now()
	if err := h.Store.Update(r.Context(), v); err != nil {
		http.Error(w, "failed to update video metadata", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(v)
}


