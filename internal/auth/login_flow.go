package auth

import (
	"context"
	"fmt"
	"net/url"
)

type LoginFlow struct {
	OAuthClient    OAuthClient
	Store          TokenStore
	BrowserOpener  func(url string) error
	RedirectReader func() (string, error)
}

func (f *LoginFlow) Run(ctx context.Context) error {
	verifier, err := NewCodeVerifier()
	if err != nil {
		return err
	}
	challenge := CodeChallengeS256(verifier)
	state, err := generateState()
	if err != nil {
		return err
	}

	authURL := f.OAuthClient.AuthorizationURL(state, challenge)
	if err := f.BrowserOpener(authURL); err != nil {
		return err
	}

	redirectURL, err := f.RedirectReader()
	if err != nil {
		return err
	}

	code, err := parseCallback(redirectURL, state)
	if err != nil {
		return err
	}

	token, err := f.OAuthClient.Exchange(ctx, code, verifier)
	if err != nil {
		return err
	}

	return f.Store.Save(ctx, token)
}

func parseCallback(rawURL, expectedState string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid redirect URL: %w", err)
	}

	query := parsed.Query()

	if errParam := query.Get("error"); errParam != "" {
		return "", fmt.Errorf("authorization failed: %s", errParam)
	}

	state := query.Get("state")
	if state != expectedState {
		return "", fmt.Errorf("state mismatch: got %q, want %q", state, expectedState)
	}

	code := query.Get("code")
	if code == "" {
		return "", fmt.Errorf("no code in redirect URL")
	}

	return code, nil
}

func generateState() (string, error) {
	return NewCodeVerifier()
}
