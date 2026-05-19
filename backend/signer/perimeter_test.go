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

// TestNginxPerimeterSimulation clones the Nginx secure_link_md5 logic to prove compatibility.
func TestNginxPerimeterSimulation(t *testing.T) {
	secret := "simulation-secret"
	baseURL := "http://localhost"
	s := NewSecureSigner(secret, baseURL)

	// Create a mock Nginx server implementing secure_link logic
	nginxMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Nginx Rewrite Logic: /hls/<token>/<expires>/<id>/<file>
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 5 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		token := parts[1]
		expiresStr := parts[2]
		videoID := parts[3]

		expires, _ := strconv.ParseInt(expiresStr, 10, 64)

		// Nginx secure_link_md5 enforcement: secret + expires + remote_addr + video_id
		// For simulation, we assume remote_addr is 127.0.0.1
		remoteAddr := "127.0.0.1"
		formula := fmt.Sprintf("%s%d%s%s", secret, expires, remoteAddr, videoID)
		hash := md5.Sum([]byte(formula))
		expectedToken := base64.RawURLEncoding.EncodeToString(hash[:])

		// 1. Expiry Check (410 Gone)
		if time.Now().Unix() > expires {
			w.WriteHeader(http.StatusGone) // 410
			return
		}

		// 2. Token Check (403 Forbidden)
		if token != expectedToken {
			w.WriteHeader(http.StatusForbidden) // 403
			return
		}

		// 3. Success (200 OK)
		w.WriteHeader(http.StatusOK)
	}))
	defer nginxMock.Close()

	videoID := "vid-123"
	remoteIP := "127.0.0.1"

	// SCENARIO 1: Valid Request (200 OK)
	signed, _ := s.Sign(videoID, remoteIP, 1*time.Hour)
	reqURL := strings.Replace(signed.URL, baseURL, nginxMock.URL, 1)
	resp, _ := http.Get(reqURL)
	if resp.StatusCode != 200 {
		t.Errorf("Simulation FAILED: Valid token rejected with %d", resp.StatusCode)
	}

	// SCENARIO 2: Expired Token (410 Gone)
	// Manual generation of an expired link
	pastExpires := time.Now().Add(-1 * time.Hour).Unix()
	expiredToken := s.GenerateToken(videoID, remoteIP, pastExpires)
	expiredURL := fmt.Sprintf("%s/hls/%s/%d/%s/master.m3u8", nginxMock.URL, expiredToken, pastExpires, videoID)
	resp, _ = http.Get(expiredURL)
	if resp.StatusCode != 410 {
		t.Errorf("Simulation FAILED: Expired token did not return 410 (Got %d)", resp.StatusCode)
	}

	// SCENARIO 3: Invalid Token (403 Forbidden)
	invalidURL := fmt.Sprintf("%s/hls/badtoken/%d/%s/master.m3u8", nginxMock.URL, signed.Expires, videoID)
	resp, _ = http.Get(invalidURL)
	if resp.StatusCode != 403 {
		t.Errorf("Simulation FAILED: Invalid token did not return 403 (Got %d)", resp.StatusCode)
	}
}
