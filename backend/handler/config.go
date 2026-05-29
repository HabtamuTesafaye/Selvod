package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

func (h *VideoHandler) HandleGetPlayerConfig(w http.ResponseWriter, r *http.Request) {
	configPath := filepath.Join(h.StorageDir, "player_config.json")
	
	h.configMutex.Lock()
	defer h.configMutex.Unlock()
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{}`))
			return
		}
		http.Error(w, "Failed to read config", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *VideoHandler) HandleUpdatePlayerConfig(w http.ResponseWriter, r *http.Request) {
	configPath := filepath.Join(h.StorageDir, "player_config.json")

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit

	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	data, err := json.Marshal(config)
	if err != nil {
		http.Error(w, "Failed to serialize config", http.StatusInternalServerError)
		return
	}
	
	h.configMutex.Lock()
	defer h.configMutex.Unlock()
	
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}
