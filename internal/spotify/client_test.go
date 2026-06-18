package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetPlaylist(t *testing.T) {
	t.Run("fetches playlist name and tracks in single page", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("method = %s, want GET", r.Method)
			}
			if auth := r.Header.Get("Authorization"); auth != "Bearer my-token" {
				t.Errorf("Authorization = %q", auth)
			}

			w.Header().Set("Content-Type", "application/json")

			if r.URL.Path == "/v1/playlists/abc123" {
				w.Write([]byte(`{
					"name": "My Playlist",
					"tracks": {
						"total": 2,
						"items": [
							{"track": {"name": "Song A", "artists": [{"name": "Artist 1"}], "album": {"name": "Album X"}, "duration_ms": 210000}},
							{"track": {"name": "Song B", "artists": [{"name": "Artist 2"}, {"name": "Artist 3"}], "album": {"name": "Album Y"}, "duration_ms": 185000}}
						]
					}
				}`))
			}
		}))
		defer server.Close()

		client := Client{
			HTTPClient: server.Client(),
			BaseURL:    server.URL,
		}

		playlist, err := client.GetPlaylist(context.Background(), "my-token", "abc123")
		if err != nil {
			t.Fatal(err)
		}

		if playlist.Name != "My Playlist" {
			t.Errorf("Name = %q", playlist.Name)
		}
		if len(playlist.Tracks) != 2 {
			t.Fatalf("len = %d, want 2", len(playlist.Tracks))
		}
		if playlist.Tracks[0].Name != "Song A" {
			t.Errorf("Tracks[0].Name = %q", playlist.Tracks[0].Name)
		}
		if playlist.Tracks[0].Artists[0] != "Artist 1" {
			t.Errorf("Tracks[0].Artists = %v", playlist.Tracks[0].Artists)
		}
		if playlist.Tracks[1].Artists[1] != "Artist 3" {
			t.Errorf("Tracks[1].Artists = %v", playlist.Tracks[1].Artists)
		}
		if playlist.Tracks[0].Album != "Album X" {
			t.Errorf("Tracks[0].Album = %q", playlist.Tracks[0].Album)
		}
		if playlist.Tracks[0].DurationMs != 210000 {
			t.Errorf("Tracks[0].DurationMs = %d", playlist.Tracks[0].DurationMs)
		}
	})

	t.Run("paginates through multiple pages", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.Header().Set("Content-Type", "application/json")

			if r.URL.Path == "/v1/playlists/playlist123" {
				w.Write([]byte(`{
					"name": "Big Playlist",
					"tracks": {
						"total": 3,
						"items": [
							{"track": {"name": "Song 1", "artists": [{"name": "A"}], "album": {"name": "Alb"}, "duration_ms": 100}},
							{"track": {"name": "Song 2", "artists": [{"name": "B"}], "album": {"name": "Alb"}, "duration_ms": 200}}
						]
					}
				}`))
			} else if r.URL.Path == "/v1/playlists/playlist123/tracks" {
				if r.URL.Query().Get("offset") != "2" {
					t.Errorf("offset = %s, want 2", r.URL.Query().Get("offset"))
				}
				w.Write([]byte(`{
					"total": 3,
					"items": [
						{"track": {"name": "Song 3", "artists": [{"name": "C"}], "album": {"name": "Alb"}, "duration_ms": 300}}
					]
				}`))
			}
		}))
		defer server.Close()

		client := Client{
			HTTPClient: server.Client(),
			BaseURL:    server.URL,
			PageSize:   2,
		}

		playlist, err := client.GetPlaylist(context.Background(), "my-token", "playlist123")
		if err != nil {
			t.Fatal(err)
		}

		if playlist.Name != "Big Playlist" {
			t.Errorf("Name = %q", playlist.Name)
		}
		if len(playlist.Tracks) != 3 {
			t.Fatalf("len = %d, want 3", len(playlist.Tracks))
		}
		if playlist.Tracks[2].Name != "Song 3" {
			t.Errorf("Tracks[2].Name = %q", playlist.Tracks[2].Name)
		}
		if callCount != 2 {
			t.Errorf("callCount = %d, want 2", callCount)
		}
	})

	t.Run("returns error on non-2xx status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client := Client{
			HTTPClient: server.Client(),
			BaseURL:    server.URL,
		}

		_, err := client.GetPlaylist(context.Background(), "bad-token", "abc123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on network failure", func(t *testing.T) {
		client := Client{
			HTTPClient: &http.Client{
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("network down")
				}),
			},
			BaseURL: "http://unreachable",
		}

		_, err := client.GetPlaylist(context.Background(), "token", "abc123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns error on invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not-json"))
		}))
		defer server.Close()

		client := Client{
			HTTPClient: server.Client(),
			BaseURL:    server.URL,
		}

		_, err := client.GetPlaylist(context.Background(), "token", "abc123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("returns ErrForbidden on 403", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		client := Client{
			HTTPClient: server.Client(),
			BaseURL:    server.URL,
		}

		_, err := client.GetPlaylist(context.Background(), "token", "abc123")
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("err = %v, want ErrForbidden", err)
		}
	})

	t.Run("retries on 429 with Retry-After header", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				w.Header().Set("Retry-After", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{
				"name": "Retry Playlist",
				"tracks": {
					"total": 1,
					"items": [
						{"track": {"name": "Song", "artists": [{"name": "A"}], "album": {"name": "B"}, "duration_ms": 100}}
					]
				}
			}`))
		}))
		defer server.Close()

		client := Client{
			HTTPClient: server.Client(),
			BaseURL:    server.URL,
		}

		playlist, err := client.GetPlaylist(context.Background(), "token", "abc123")
		if err != nil {
			t.Fatal(err)
		}
		if playlist.Name != "Retry Playlist" {
			t.Errorf("Name = %q", playlist.Name)
		}
		if callCount != 2 {
			t.Errorf("callCount = %d, want 2", callCount)
		}
	})

	t.Run("fails after retry exhausted on 429", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client := Client{
			HTTPClient: server.Client(),
			BaseURL:    server.URL,
		}

		_, err := client.GetPlaylist(context.Background(), "token", "abc123")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
