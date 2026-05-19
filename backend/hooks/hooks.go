package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/selvod/selvod/store"
)

// Event represents a specific lifecycle stage of a video.
type Event string

const (
	EventUpload            Event = "video.upload"
	EventTranscodeStart    Event = "video.transcoding"
	EventTranscodeComplete Event = "video.completed"
	EventDelete            Event = "video.deleted"
)

// Hook is the interface for responding to system events.
type Hook interface {
	OnEvent(event Event, video *store.Video) error
}

// WebhookHook sends a JSON POST request to a configured URL.
type WebhookHook struct {
	URL    string
	Client *http.Client
}

func NewWebhookHook(url string) *WebhookHook {
	return &WebhookHook{
		URL:    url,
		Client: &http.Client{Timeout: 5 * time.Second},
	}
}

// OnEvent delivers the video metadata to the webhook endpoint.
func (h *WebhookHook) OnEvent(event Event, video *store.Video) error {
	if h.URL == "" {
		return nil
	}

	payload := map[string]interface{}{
		"event":     event,
		"timestamp": time.Now(),
		"video":     video,
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", h.URL, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Registry manages a collection of hooks.
type Registry struct {
	hooks []Hook
}

func NewRegistry() *Registry {
	return &Registry{hooks: []Hook{}}
}

func (r *Registry) Add(h Hook) {
	r.hooks = append(r.hooks, h)
}

// Dispatch broadcasts an event to all registered hooks.
func (r *Registry) Dispatch(event Event, video *store.Video) {
	for _, h := range r.hooks {
		_ = h.OnEvent(event, video)
	}
}
