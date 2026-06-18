package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestRequestToken(t *testing.T) {
	fixedNow := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	t.Run("posts form values and returns token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
				t.Errorf("content-type = %s", ct)
			}

			r.ParseForm()
			if r.Form.Get("foo") != "bar" {
				t.Errorf("foo = %s", r.Form.Get("foo"))
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"access_token": "at",
				"token_type": "Bearer",
				"scope": "read",
				"expires_in": 1800,
				"refresh_token": "rt"
			}`))
		}))
		defer server.Close()

		client := &http.Client{Transport: rewriteTo(server)}
		values := url.Values{}
		values.Set("foo", "bar")

		got, err := requestToken(context.Background(), client, func() time.Time { return fixedNow }, values)
		if err != nil {
			t.Fatal(err)
		}

		if got.AccessToken != "at" {
			t.Errorf("AccessToken = %q", got.AccessToken)
		}
		if got.RefreshToken != "rt" {
			t.Errorf("RefreshToken = %q", got.RefreshToken)
		}
		if got.TokenType != "Bearer" {
			t.Errorf("TokenType = %q", got.TokenType)
		}
		if got.Scope != "read" {
			t.Errorf("Scope = %q", got.Scope)
		}
		if want := fixedNow.Add(1800 * time.Second); !got.Expiry.Equal(want) {
			t.Errorf("Expiry = %v, want %v", got.Expiry, want)
		}
	})

	t.Run("returns error on non-2xx", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := &http.Client{Transport: rewriteTo(server)}

		_, err := requestToken(context.Background(), client, func() time.Time { return fixedNow }, url.Values{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("nope"))
		}))
		defer server.Close()

		client := &http.Client{Transport: rewriteTo(server)}

		_, err := requestToken(context.Background(), client, func() time.Time { return fixedNow }, url.Values{})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("uses default client and time.Now when nil", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
		}))
		defer server.Close()

		original := http.DefaultTransport
		http.DefaultTransport = rewriteTo(server)
		t.Cleanup(func() { http.DefaultTransport = original })

		got, err := requestToken(context.Background(), nil, nil, url.Values{})
		if err != nil {
			t.Fatal(err)
		}
		if got.AccessToken != "tok" {
			t.Errorf("AccessToken = %q", got.AccessToken)
		}
		if got.Expiry.IsZero() {
			t.Error("Expiry should not be zero")
		}
	})
}
