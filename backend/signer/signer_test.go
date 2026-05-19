package signer

import (
	"testing"
	"time"
)

func TestSigner(t *testing.T) {
	s := NewSecureSigner("secret", "http://localhost")
	
	// Test basic signing
	signed, err := s.Sign("video-1", "127.0.0.1", 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if signed.Token == "" {
		t.Error("expected non-empty token")
	}

	// Test IP binding (Tokens must differ for different IPs)
	signedOther, _ := s.Sign("video-1", "1.1.1.1", 1*time.Hour)
	if signed.Token == signedOther.Token {
		t.Error("tokens should be unique per IP")
	}
}
