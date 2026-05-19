package signer

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNginxPerimeterSimulation(t *testing.T) {
	globalSecret := "global-secret"
	playbackSecret := "lib-secret"
	baseURL := "http://localhost"
	s := NewSecureSigner(globalSecret, baseURL)

	// Create a mock Nginx server implementing dynamic validation simulation
	nginxMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Path: /hls/{library_id}/{video_id}/master.m3u8
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		libraryID := parts[1]
		videoID := parts[2]

		token := r.URL.Query().Get("token")
		expiresStr := r.URL.Query().Get("expires")
		expires, _ := strconv.ParseInt(expiresStr, 10, 64)

		// Re-simulate MD5 formula: secret + "/" + library_id + "/" + video_id + "/" + expires
		formula := fmt.Sprintf("%s/%s/%s/%d", playbackSecret, libraryID, videoID, expires)
		hash := md5.Sum([]byte(formula))
		expectedToken := base64.RawURLEncoding.EncodeToString(hash[:])

		// 1. Expiry Check
		if time.Now().Unix() > expires {
			w.WriteHeader(http.StatusGone) // 410
			return
		}

		// 2. Token Check
		if token != expectedToken {
			w.WriteHeader(http.StatusForbidden) // 403
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer nginxMock.Close()

	libraryID := "lib-123"
	videoID := "vid-123"

	// SCENARIO 1: Valid Request (200 OK)
	signed, _ := s.Sign(libraryID, videoID, playbackSecret, 1*time.Hour)
	reqURL := strings.Replace(signed.URL, baseURL, nginxMock.URL, 1)
	resp, _ := http.Get(reqURL)
	if resp.StatusCode != 200 {
		t.Errorf("Simulation FAILED: Valid token rejected with %d", resp.StatusCode)
	}

	// SCENARIO 2: Expired Token (410 Gone)
	pastExpires := time.Now().Add(-1 * time.Hour).Unix()
	expiredToken := s.GenerateToken(libraryID, videoID, playbackSecret, pastExpires)
	expiredURL := fmt.Sprintf("%s/hls/%s/%s/master.m3u8?token=%s&expires=%d", nginxMock.URL, libraryID, videoID, expiredToken, pastExpires)
	resp, _ = http.Get(expiredURL)
	if resp.StatusCode != 410 {
		t.Errorf("Simulation FAILED: Expired token did not return 410 (Got %d)", resp.StatusCode)
	}

	// SCENARIO 3: Invalid Token (403 Forbidden)
	invalidURL := fmt.Sprintf("%s/hls/%s/%s/master.m3u8?token=badtoken&expires=%d", nginxMock.URL, libraryID, videoID, signed.Expires)
	resp, _ = http.Get(invalidURL)
	if resp.StatusCode != 403 {
		t.Errorf("Simulation FAILED: Invalid token did not return 403 (Got %d)", resp.StatusCode)
	}
}
