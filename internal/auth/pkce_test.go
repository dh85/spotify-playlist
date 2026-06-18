package auth

import (
	"encoding/base64"
	"regexp"
	"testing"
	"testing/iotest"
)

func TestNewCodeVerifier(t *testing.T) {
	t.Run("has valid PKCE length and characters", func(t *testing.T) {
		verifier, err := NewCodeVerifier()
		if err != nil {
			t.Fatal(err)
		}

		if len(verifier) < 43 || len(verifier) > 128 {
			t.Errorf("verifier length = %d, want 43..128", len(verifier))
		}

		allowed := regexp.MustCompile(`^[A-Za-z0-9._~-]+$`)
		if !allowed.MatchString(verifier) {
			t.Fatalf("verifier contains invalid characters: %q", verifier)
		}
	})

	t.Run("returns error when rand fails", func(t *testing.T) {
		_, err := NewCodeVerifierFrom(iotest.ErrReader(iotest.ErrTimeout))
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestCodeChallengeS256(t *testing.T) {
	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"

	got := CodeChallengeS256(verifier)

	want := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
	if got != want {
		t.Fatalf("challenge = %q, want %q", got, want)
	}

	if _, err := base64.RawURLEncoding.DecodeString(got); err != nil {
		t.Fatalf("challenge should be raw base64url: %v", err)
	}
}
