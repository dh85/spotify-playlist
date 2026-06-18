package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const tokenEndpoint = "https://accounts.spotify.com/api/token"

func requestToken(ctx context.Context, client *http.Client, now func() time.Time, values url.Values) (Token, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		tokenEndpoint,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return Token{}, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Token{}, fmt.Errorf("token request failed: status %d: %s", resp.StatusCode, string(body))
	}

	var body struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Token{}, err
	}

	if now == nil {
		now = time.Now
	}

	return Token{
		AccessToken:  body.AccessToken,
		RefreshToken: body.RefreshToken,
		TokenType:    body.TokenType,
		Scope:        body.Scope,
		Expiry:       now().Add(time.Duration(body.ExpiresIn) * time.Second),
	}, nil
}
