package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSpotifyTokenRefresher(t *testing.T) {
	fixedNow := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	t.Run("returns token on success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/x-www-form-urlencoded" {
				t.Errorf("content-type = %s", ct)
			}

			r.ParseForm()
			if r.Form.Get("grant_type") != "refresh_token" {
				t.Errorf("grant_type = %s", r.Form.Get("grant_type"))
			}
			if r.Form.Get("refresh_token") != "my-refresh" {
				t.Errorf("refresh_token = %s", r.Form.Get("refresh_token"))
			}
			if r.Form.Get("client_id") != "my-client" {
				t.Errorf("client_id = %s", r.Form.Get("client_id"))
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

		refresher := SpotifyTokenRefresher{
			ClientID:   "my-client",
			HTTPClient: &http.Client{Transport: rewriteTo(server)},
			Now:        func() time.Time { return fixedNow },
		}

		got, err := refresher.Refresh(context.Background(), "my-refresh")
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
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		refresher := SpotifyTokenRefresher{
			ClientID:   "my-client",
			HTTPClient: &http.Client{Transport: rewriteTo(server)},
			Now:        func() time.Time { return fixedNow },
		}

		_, err := refresher.Refresh(context.Background(), "bad-refresh")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on network failure", func(t *testing.T) {
		refresher := SpotifyTokenRefresher{
			ClientID: "my-client",
			HTTPClient: &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return nil, context.DeadlineExceeded
				}),
			},
			Now: func() time.Time { return fixedNow },
		}

		_, err := refresher.Refresh(context.Background(), "token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not-json"))
		}))
		defer server.Close()

		refresher := SpotifyTokenRefresher{
			ClientID:   "my-client",
			HTTPClient: &http.Client{Transport: rewriteTo(server)},
			Now:        func() time.Time { return fixedNow },
		}

		_, err := refresher.Refresh(context.Background(), "token")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("uses defaults for nil HTTPClient and Now", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
		}))
		defer server.Close()

		original := http.DefaultTransport
		http.DefaultTransport = rewriteTo(server)
		t.Cleanup(func() { http.DefaultTransport = original })

		refresher := SpotifyTokenRefresher{
			ClientID: "my-client",
		}

		got, err := refresher.Refresh(context.Background(), "token")
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// rewriteTo redirects requests to the test server using the real http transport.
func rewriteTo(server *httptest.Server) http.RoundTripper {
	transport := &http.Transport{}
	return roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = server.Listener.Addr().String()
		return transport.RoundTrip(req)
	})
}
