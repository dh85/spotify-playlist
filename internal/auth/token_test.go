package auth

import (
	"testing"
	"time"
)

func TestIsValidReturnsFalseWhenAccessTokenIsEmpty(t *testing.T) {
	token := Token{
		AccessToken: "",
		Expiry:      time.Now().Add(time.Hour),
	}

	if token.IsValid(time.Now()) {
		t.Fatal("expected false for empty access token")
	}
}
