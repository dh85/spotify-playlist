package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthorizationURL(t *testing.T) {
	c := OAuthClient{
		ClientID:    "my-client-id",
		RedirectURI: "http://localhost:8080/callback",
		Scopes:      []string{"user-read-private", "playlist-modify-public"},
	}

	got := c.AuthorizationURL("abc123", "challenge456")

	want := "https://accounts.spotify.com/authorize?" +
		"client_id=my-client-id" +
		"&code_challenge=challenge456" +
		"&code_challenge_method=S256" +
		"&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback" +
		"&response_type=code" +
		"&scope=user-read-private+playlist-modify-public" +
		"&state=abc123"

	if got != want {
		t.Errorf("got:\n  %s\nwant:\n  %s", got, want)
	}
}

func TestExchange(t *testing.T) {
	fixedNow := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	t.Run("exchanges code for token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
				t.Errorf("content-type = %s", ct)
			}

			r.ParseForm()
			if r.Form.Get("grant_type") != "authorization_code" {
				t.Errorf("grant_type = %s", r.Form.Get("grant_type"))
			}
			if r.Form.Get("code") != "my-auth-code" {
				t.Errorf("code = %s", r.Form.Get("code"))
			}
			if r.Form.Get("redirect_uri") != "http://localhost:8080/callback" {
				t.Errorf("redirect_uri = %s", r.Form.Get("redirect_uri"))
			}
			if r.Form.Get("client_id") != "my-client" {
				t.Errorf("client_id = %s", r.Form.Get("client_id"))
			}
			if r.Form.Get("code_verifier") != "my-verifier" {
				t.Errorf("code_verifier = %s", r.Form.Get("code_verifier"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"access_token": "new-access",
				"token_type": "Bearer",
				"scope": "user-read-private",
				"expires_in": 3600,
				"refresh_token": "new-refresh"
			}`))
		}))
		defer server.Close()

		client := OAuthClient{
			ClientID:    "my-client",
			RedirectURI: "http://localhost:8080/callback",
			HTTPClient:  &http.Client{Transport: rewriteTo(server)},
			Now:         func() time.Time { return fixedNow },
		}

		got, err := client.Exchange(context.Background(), "my-auth-code", "my-verifier")
		if err != nil {
			t.Fatal(err)
		}

		if got.AccessToken != "new-access" {
			t.Errorf("AccessToken = %q", got.AccessToken)
		}
		if got.RefreshToken != "new-refresh" {
			t.Errorf("RefreshToken = %q", got.RefreshToken)
		}
		if got.TokenType != "Bearer" {
			t.Errorf("TokenType = %q", got.TokenType)
		}
		if got.Scope != "user-read-private" {
			t.Errorf("Scope = %q", got.Scope)
		}
		if want := fixedNow.Add(3600 * time.Second); !got.Expiry.Equal(want) {
			t.Errorf("Expiry = %v, want %v", got.Expiry, want)
		}
	})

	t.Run("returns error on non-2xx status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := OAuthClient{
			ClientID:    "my-client",
			RedirectURI: "http://localhost:8080/callback",
			HTTPClient:  &http.Client{Transport: rewriteTo(server)},
			Now:         func() time.Time { return fixedNow },
		}

		_, err := client.Exchange(context.Background(), "bad-code", "verifier")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on network failure", func(t *testing.T) {
		client := OAuthClient{
			ClientID:    "my-client",
			RedirectURI: "http://localhost:8080/callback",
			HTTPClient: &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return nil, context.DeadlineExceeded
				}),
			},
			Now: func() time.Time { return fixedNow },
		}

		_, err := client.Exchange(context.Background(), "code", "verifier")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not-json"))
		}))
		defer server.Close()

		client := OAuthClient{
			ClientID:    "my-client",
			RedirectURI: "http://localhost:8080/callback",
			HTTPClient:  &http.Client{Transport: rewriteTo(server)},
			Now:         func() time.Time { return fixedNow },
		}

		_, err := client.Exchange(context.Background(), "code", "verifier")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
