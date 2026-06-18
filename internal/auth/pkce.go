package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
)

// verifierBytes is the number of random bytes used to generate a code verifier.
// 64 bytes encode to 86 base64url chars, safely within PKCE's 43..128 range.
const verifierBytes = 64

func NewCodeVerifier() (string, error) {
	return NewCodeVerifierFrom(rand.Reader)
}

func NewCodeVerifierFrom(reader io.Reader) (string, error) {
	bytes := make([]byte, verifierBytes)
	if _, err := io.ReadFull(reader, bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func CodeChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
