package signer

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestIndependentNginxCompatibility(t *testing.T) {
	// GIVEN: A known context
	secret := "test-secret"
	videoID := "vid-123"
	remoteIP := "127.0.0.1"
	
	s := NewSecureSigner(secret, "http://localhost")
	signed, _ := s.Sign(videoID, remoteIP, 1*time.Hour)

	// WHEN: We MANUALLY re-implement the Nginx formula (independent of the signer package)
	// Logic: secret + expires + remote_addr + video_id
	formula := fmt.Sprintf("%s%d%s%s", secret, signed.Expires, remoteIP, videoID)
	hash := md5.Sum([]byte(formula))
	
	// Nginx uses RawURLEncoding (no padding, URL-safe characters)
	expectedToken := base64.RawURLEncoding.EncodeToString(hash[:])

	// THEN: The backend-generated token must match the manually re-implemented Nginx token
	if signed.Token != expectedToken {
		t.Errorf("CRITICAL INCOMPATIBILITY: Backend token (%s) does not match Nginx formula (%s)", signed.Token, expectedToken)
	}

	// Verify the URL structure follows the Nginx rewrite requirement
	// Format: /hls/<token>/<expires>/<id>/master.m3u8
	requiredPath := fmt.Sprintf("/hls/%s/%d/%s/", signed.Token, signed.Expires, videoID)
	if !strings.Contains(signed.URL, requiredPath) {
		t.Errorf("URL structure mismatch! Required path part: %s, Got: %s", requiredPath, signed.URL)
	}
}
