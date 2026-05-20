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
	globalSecret := "global-secret"
	playbackSecret := "lib-secret"
	videoID := "vid-123"
	libraryID := "lib-abc"
	
	s := NewSecureSigner(globalSecret, "http://localhost")
	signed, _ := s.Sign(libraryID, videoID, playbackSecret, 1*time.Hour)

	// WHEN: We MANUALLY re-implement the Nginx/Go multi-library formula
	// Logic: secret + "/" + library_id + "/" + video_id + "/" + expires
	formula := fmt.Sprintf("%s/%s/%s/%d", playbackSecret, libraryID, videoID, signed.Expires)
	hash := md5.Sum([]byte(formula))
	
	// Nginx uses RawURLEncoding (no padding, URL-safe characters)
	expectedToken := base64.RawURLEncoding.EncodeToString(hash[:])

	// THEN: The backend-generated token must match the manually re-implemented Nginx token
	if signed.Token != expectedToken {
		t.Errorf("CRITICAL INCOMPATIBILITY: Backend token (%s) does not match Nginx formula (%s)", signed.Token, expectedToken)
	}

	// Verify the URL structure follows the new query param structure
	// Format: /hls/<library_id>/<video_id>/master.m3u8?token=<token>&expires=<expires>
	requiredPath := fmt.Sprintf("/hls/%s/%s/master.m3u8?token=%s&expires=%d", libraryID, videoID, signed.Token, signed.Expires)
	if !strings.Contains(signed.URL, requiredPath) {
		t.Errorf("URL structure mismatch! Required path part: %s, Got: %s", requiredPath, signed.URL)
	}
}
