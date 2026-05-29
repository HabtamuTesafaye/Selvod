package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/selvod/selvod/store"
)

type LibraryCache struct {
	mu      sync.RWMutex
	secrets map[string]cachedSecret
}

type cachedSecret struct {
	secret    string
	expiresAt time.Time
}

func NewLibraryCache() *LibraryCache {
	return &LibraryCache{
		secrets: make(map[string]cachedSecret),
	}
}

func (c *LibraryCache) Get(libraryID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	cached, ok := c.secrets[libraryID]
	if !ok || time.Now().After(cached.expiresAt) {
		return "", false
	}
	return cached.secret, true
}

func (c *LibraryCache) Set(libraryID string, secret string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.secrets[libraryID] = cachedSecret{
		secret:    secret,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *LibraryCache) Evict(libraryID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.secrets, libraryID)
}

// Library Handlers

func (h *VideoHandler) HandleListLibraries(w http.ResponseWriter, r *http.Request) {
	libs, err := h.Store.ListLibraries(r.Context())
	if err != nil {
		http.Error(w, "failed to list libraries", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(libs)
}

func (h *VideoHandler) HandleCreateLibrary(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	lib := &store.Library{
		ID:   uuid.New().String(),
		Name: input.Name,
	}

	if err := h.Store.CreateLibrary(r.Context(), lib); err != nil {
		http.Error(w, "failed to create library", http.StatusInternalServerError)
		return
	}

	var defaultKey *store.LibraryKey

	// Auto-generate a default playback key for the new library
	keyBytes := make([]byte, 16)
	if _, err := rand.Read(keyBytes); err == nil {
		secret := hex.EncodeToString(keyBytes)
		defaultKey = &store.LibraryKey{
			ID:             uuid.New().String(),
			LibraryID:      lib.ID,
			KeyName:        "Default Playback Key",
			PlaybackSecret: secret,
			IsActive:       true,
		}
		if err := h.Store.CreateLibraryKey(r.Context(), defaultKey); err != nil {
			slog.Warn("failed to create default key", "library_id", lib.ID, "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	// Return both the library AND the generated key secret (only shown once)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"library":     lib,
		"default_key": defaultKey,
	})
}

func (h *VideoHandler) HandleUpdateLibrary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lib, err := h.Store.GetLibrary(r.Context(), id)
	if err != nil || lib == nil {
		http.Error(w, "library not found", http.StatusNotFound)
		return
	}

	var input struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Name == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	lib.Name = input.Name
	if err := h.Store.UpdateLibrary(r.Context(), lib); err != nil {
		http.Error(w, "failed to update library name", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lib)
}

func (h *VideoHandler) HandleListLibraryKeys(w http.ResponseWriter, r *http.Request) {
	libraryID := chi.URLParam(r, "id")
	keys, err := h.Store.ListLibraryKeys(r.Context(), libraryID)
	if err != nil {
		http.Error(w, "failed to list library keys", http.StatusInternalServerError)
		return
	}

	sanitized := make([]map[string]interface{}, 0, len(keys))
	for _, k := range keys {
		sanitized = append(sanitized, map[string]interface{}{
			"id":         k.ID,
			"library_id": k.LibraryID,
			"key_name":   k.KeyName,
			"is_active":  k.IsActive,
			"created_at": k.CreatedAt,
		})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sanitized)
}

func (h *VideoHandler) HandleCreateLibraryKey(w http.ResponseWriter, r *http.Request) {
	libraryID := chi.URLParam(r, "id")
	var input struct {
		KeyName string `json:"key_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.KeyName == "" {
		http.Error(w, "invalid input", http.StatusBadRequest)
		return
	}

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		http.Error(w, "failed to generate key secret", http.StatusInternalServerError)
		return
	}
	secret := hex.EncodeToString(bytes)

	k := &store.LibraryKey{
		ID:             uuid.New().String(),
		LibraryID:      libraryID,
		KeyName:        input.KeyName,
		PlaybackSecret: secret,
	}

	if err := h.Store.CreateLibraryKey(r.Context(), k); err != nil {
		http.Error(w, "failed to save library key", http.StatusInternalServerError)
		return
	}

	// Evict cache to force reload of keys
	h.Cache.Evict(libraryID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(k)
}

func (h *VideoHandler) HandleRevokeLibraryKey(w http.ResponseWriter, r *http.Request) {
	libraryID := chi.URLParam(r, "id")
	keyID := chi.URLParam(r, "key_id")

	if err := h.Store.RevokeLibraryKey(r.Context(), keyID); err != nil {
		http.Error(w, "failed to revoke key", http.StatusInternalServerError)
		return
	}

	// Immediate Cache Eviction on key revocation
	h.Cache.Evict(libraryID)

	w.WriteHeader(http.StatusNoContent)
}

func (h *VideoHandler) HandleRegenerateLibraryKey(w http.ResponseWriter, r *http.Request) {
	libraryID := chi.URLParam(r, "id")
	keyID := chi.URLParam(r, "key_id")

	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		http.Error(w, "failed to generate key secret", http.StatusInternalServerError)
		return
	}
	secret := hex.EncodeToString(bytes)

	if err := h.Store.RegenerateLibraryKey(r.Context(), keyID, secret); err != nil {
		http.Error(w, "failed to regenerate key", http.StatusInternalServerError)
		return
	}

	// Immediate Cache Eviction on key modification
	h.Cache.Evict(libraryID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"playback_secret": secret,
	})
}

func (h *VideoHandler) HandleDeleteLibraryKey(w http.ResponseWriter, r *http.Request) {
	libraryID := chi.URLParam(r, "id")
	keyID := chi.URLParam(r, "key_id")

	if err := h.Store.DeleteLibraryKey(r.Context(), keyID); err != nil {
		http.Error(w, "failed to delete key", http.StatusInternalServerError)
		return
	}

	// Immediate Cache Eviction
	h.Cache.Evict(libraryID)

	w.WriteHeader(http.StatusNoContent)
}

// Authentication Endpoint for Nginx auth_request

func (h *VideoHandler) HandleAuthManifest(w http.ResponseWriter, r *http.Request) {
	origURI := r.Header.Get("X-Original-URI")
	if origURI == "" {
		origURI = r.URL.RequestURI()
	}

	u, err := url.Parse(origURI)
	if err != nil {
		http.Error(w, "invalid original URI", http.StatusBadRequest)
		return
	}

	// Path: /hls/{library_id}/{video_id}/{filename}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 3 || parts[0] != "hls" {
		http.Error(w, "malformed original path", http.StatusBadRequest)
		return
	}
	libraryID := parts[1]
	videoID := parts[2]

	token := u.Query().Get("token")
	expiresStr := u.Query().Get("expires")
	if token == "" || expiresStr == "" {
		http.Error(w, "missing signature parameters", http.StatusUnauthorized)
		return
	}

	expires, err := strconv.ParseInt(expiresStr, 10, 64)
	if err != nil || time.Now().Unix() > expires {
		w.Header().Set("X-Auth-Status", "410")
		http.Error(w, "signature expired", http.StatusGone)
		return
	}

	// Retrieve playback secret (Cache with TTL / DB fallback)
	secret, err := h.playbackSecret(r.Context(), libraryID)
	if err != nil {
		http.Error(w, "no active playback keys found for library", http.StatusUnauthorized)
		return
	}

	// Verify token
	expected := h.Signer.GenerateToken(libraryID, videoID, secret, expires)
	if token != expected {
		http.Error(w, "invalid signature token", http.StatusForbidden)
		return
	}

	// Generate segment validation cookie
	segmentToken := h.Signer.GenerateSegmentToken(videoID, expires)
	cookieValue := fmt.Sprintf("%s,%d", segmentToken, expires)
	cookieName := fmt.Sprintf("sv_session_%s", videoID)
	cookiePath := fmt.Sprintf("/hls/%s/%s/", libraryID, videoID)

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    cookieValue,
		Path:     cookiePath,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		Expires:  time.Unix(expires, 0),
	}
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
}
