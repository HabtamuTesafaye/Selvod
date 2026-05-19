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
	secret  string
	baseURL string
}

func NewSecureSigner(secret, baseURL string) *SecureSigner {
	return &SecureSigner{
		secret:  secret,
		baseURL: baseURL,
	}
}

// Sign generates a cryptographically bound URL using MD5 + IP Binding + Expiry.
func (s *SecureSigner) Sign(videoID string, remoteIP string, duration time.Duration) (*SignedURL, error) {
	expires := time.Now().Add(duration).Unix()
	token := s.GenerateToken(videoID, remoteIP, expires)

	fullURL := fmt.Sprintf("%s/hls/%s/%d/%s/master.m3u8", s.baseURL, token, expires, videoID)

	return &SignedURL{
		URL:       fullURL,
		Token:     token,
		Expires:   expires,
		ExpiresIn: int64(duration.Seconds()),
	}, nil
}

// GenerateToken provides the raw MD5 hash for a given context.
// This is used for both generation and verification.
func (s *SecureSigner) GenerateToken(videoID string, remoteIP string, expires int64) string {
	data := fmt.Sprintf("%s%d%s%s", s.secret, expires, remoteIP, videoID)
	hash := md5.Sum([]byte(data))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
