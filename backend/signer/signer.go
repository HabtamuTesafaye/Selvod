package signer

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"time"
)

type SignedURL struct {
	URL       string `json:"url"`
	Token     string `json:"token"`
	Expires   int64  `json:"expires"`
	ExpiresIn int64  `json:"expires_in"`
}

type SecureSigner struct {
	secret  string // Global secret (for segment validation)
	baseURL string
}

func NewSecureSigner(secret, baseURL string) *SecureSigner {
	return &SecureSigner{
		secret:  secret,
		baseURL: baseURL,
	}
}

// Sign generates a cryptographically bound URL using MD5 + Library Scoped Secret + Expiry.
func (s *SecureSigner) Sign(libraryID, videoID string, playbackSecret string, duration time.Duration) (*SignedURL, error) {
	expires := time.Now().Add(duration).Unix()
	token := s.GenerateToken(libraryID, videoID, playbackSecret, expires)

	fullURL := fmt.Sprintf("%s/hls/%s/%s/master.m3u8?token=%s&expires=%d", s.baseURL, libraryID, videoID, token, expires)

	return &SignedURL{
		URL:       fullURL,
		Token:     token,
		Expires:   expires,
		ExpiresIn: int64(duration.Seconds()),
	}, nil
}

// GenerateToken provides the raw MD5 hash for a given context using the library secret.
func (s *SecureSigner) GenerateToken(libraryID, videoID string, playbackSecret string, expires int64) string {
	data := fmt.Sprintf("%s/%s/%s/%d", playbackSecret, libraryID, videoID, expires)
	hash := md5.Sum([]byte(data))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GenerateSegmentToken generates the token for segment-level cookie validation using global secret.
func (s *SecureSigner) GenerateSegmentToken(videoID string, expires int64) string {
	data := fmt.Sprintf("%s/%s/%d", s.secret, videoID, expires)
	hash := md5.Sum([]byte(data))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
