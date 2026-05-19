package signer

import (
	"testing"
	"time"
)

func TestSigner(t *testing.T) {
	s := NewSecureSigner("global-secret", "http://localhost")
	
	// Test basic signing with library secret
	signed, err := s.Sign("library-1", "video-1", "lib-secret", 1*time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	if signed.Token == "" {
		t.Error("expected non-empty token")
	}

	// Test dynamic key variations
	signedOther, _ := s.Sign("library-1", "video-1", "lib-secret-diff", 1*time.Hour)
	if signed.Token == signedOther.Token {
		t.Error("tokens should be unique per secret key")
	}
}
