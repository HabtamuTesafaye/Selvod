package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
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
	Store      store.Store
	Storage    storage.Provider
	Signer     *signer.SecureSigner
	Queue      *queue.WorkerPool
	Hooks      *hooks.Registry
	Transcoder transcoder.Transcoder
	StorageDir string
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
	title := r.FormValue("title")
	if title == "" {
		title = header.Filename
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
	if ok, _ := h.Transcoder.IsVideo(r.Context(), fullPath); !ok {
		h.Storage.Delete(r.Context(), uploadPath)
		http.Error(w, "Security rejection: Invalid or malicious media file", http.StatusUnsupportedMediaType)
		return
	}

	v := &store.Video{
		ID:              id,
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

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(v)
}

func (h *VideoHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	videos, total, err := h.Store.List(r.Context(), 50, 0)
	if err != nil {
		http.Error(w, "failed to fetch video list", http.StatusInternalServerError)
		return
	}
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

	_ = h.Storage.Delete(r.Context(), filepath.Join("uploads", id+v.OriginalExt))
	_ = h.Storage.Delete(r.Context(), filepath.Join("hls", id))

	if err := h.Store.Delete(r.Context(), id); err != nil {
		http.Error(w, "failed to remove metadata record", http.StatusInternalServerError)
		return
	}
	h.Hooks.Dispatch(hooks.EventDelete, v)
	w.WriteHeader(http.StatusNoContent)
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

	remoteIP := getRemoteIP(r)
	signed, err := h.Signer.Sign(id, remoteIP, 2*time.Hour)
	if err != nil {
		http.Error(w, "failed to generate signed URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	json.NewEncoder(w).Encode(signed)
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

func getRemoteIP(r *http.Request) string {
	remoteIP := r.Header.Get("X-Real-IP")
	if remoteIP == "" {
		remoteIP = r.Header.Get("X-Forwarded-For")
	}
	if remoteIP == "" {
		remoteIP = r.RemoteAddr
		if strings.Contains(remoteIP, ":") {
			remoteIP = strings.Split(remoteIP, ":")[0]
		}
	}
	return remoteIP
}
