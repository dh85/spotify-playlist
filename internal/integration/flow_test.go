package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dh85/spotify-playlist/internal/auth"
	"github.com/dh85/spotify-playlist/internal/spotify"
	"github.com/dh85/spotify-playlist/internal/storage"
)

func TestFullFlow(t *testing.T) {
	fixedNow := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)

	// Fake Spotify server: handles token exchange, token refresh, and playlist API
	spotifyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.URL.Path == "/api/token" && r.Method == http.MethodPost:
			r.ParseForm()
			grantType := r.Form.Get("grant_type")

			switch grantType {
			case "authorization_code":
				if r.Form.Get("code") == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				w.Write([]byte(`{
					"access_token": "initial-access-token",
					"token_type": "Bearer",
					"scope": "playlist-read-private",
					"expires_in": 3600,
					"refresh_token": "initial-refresh-token"
				}`))
			case "refresh_token":
				if r.Form.Get("refresh_token") != "initial-refresh-token" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				w.Write([]byte(`{
					"access_token": "refreshed-access-token",
					"token_type": "Bearer",
					"scope": "playlist-read-private",
					"expires_in": 3600,
					"refresh_token": ""
				}`))
			default:
				w.WriteHeader(http.StatusBadRequest)
			}

		case r.URL.Path == "/v1/playlists/test-playlist-123" && r.Method == http.MethodGet:
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer initial-access-token" && authHeader != "Bearer refreshed-access-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Write([]byte(`{
				"name": "Test Playlist",
				"tracks": {
					"total": 3,
					"items": [
						{"track": {"name": "First Song", "artists": [{"name": "Artist A"}], "album": {"name": "Album 1"}, "duration_ms": 180000}},
						{"track": {"name": "Second Song", "artists": [{"name": "Artist B"}, {"name": "Artist C"}], "album": {"name": "Album 2"}, "duration_ms": 240000}},
						{"track": {"name": "Third Song", "artists": [{"name": "Artist D"}], "album": {"name": "Album 3"}, "duration_ms": 300000}}
					]
				}
			}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer spotifyServer.Close()

	// Transport that rewrites requests to our fake server
	transport := &http.Transport{}
	rewriteClient := &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			req.URL.Scheme = "http"
			req.URL.Host = spotifyServer.Listener.Addr().String()
			return transport.RoundTrip(req)
		}),
	}

	// Shared components
	store := storage.NewMemoryTokenStore()
	ctx := context.Background()

	// --- Step 1: Login Flow ---
	t.Run("login flow stores token", func(t *testing.T) {
		var browserURL string
		flow := auth.LoginFlow{
			OAuthClient: auth.OAuthClient{
				ClientID:    "test-client-id",
				RedirectURI: "https://google.com/",
				Scopes:      []string{"playlist-read-private"},
				HTTPClient:  rewriteClient,
				Now:         func() time.Time { return fixedNow },
			},
			Store: store,
			BrowserOpener: func(u string) error {
				browserURL = u
				return nil
			},
			RedirectReader: func() (string, error) {
				// Simulate user pasting the redirect URL
				parsed, _ := url.Parse(browserURL)
				state := parsed.Query().Get("state")
				return "https://google.com/?code=auth-code-123&state=" + state, nil
			},
		}

		if err := flow.Run(ctx); err != nil {
			t.Fatalf("login flow failed: %v", err)
		}

		if browserURL == "" {
			t.Fatal("browser was never opened")
		}

		// Verify token was stored
		token, err := store.Load(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if token.AccessToken != "initial-access-token" {
			t.Errorf("AccessToken = %q", token.AccessToken)
		}
		if token.RefreshToken != "initial-refresh-token" {
			t.Errorf("RefreshToken = %q", token.RefreshToken)
		}
	})

	// --- Step 2: TokenManager returns valid token ---
	t.Run("token manager returns stored token when valid", func(t *testing.T) {
		refresher := auth.SpotifyTokenRefresher{
			ClientID:   "test-client-id",
			HTTPClient: rewriteClient,
			Now:        func() time.Time { return fixedNow },
		}

		manager := auth.NewTokenManager(store, &refresher, func() time.Time { return fixedNow })

		token, err := manager.ValidToken(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if token.AccessToken != "initial-access-token" {
			t.Errorf("AccessToken = %q", token.AccessToken)
		}
	})

	// --- Step 3: TokenManager refreshes expired token ---
	t.Run("token manager refreshes expired token", func(t *testing.T) {
		expiredNow := fixedNow.Add(2 * time.Hour) // token expired

		refresher := auth.SpotifyTokenRefresher{
			ClientID:   "test-client-id",
			HTTPClient: rewriteClient,
			Now:        func() time.Time { return expiredNow },
		}

		manager := auth.NewTokenManager(store, &refresher, func() time.Time { return expiredNow })

		token, err := manager.ValidToken(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if token.AccessToken != "refreshed-access-token" {
			t.Errorf("AccessToken = %q", token.AccessToken)
		}
		// Original refresh token should be preserved since the refresh response returned empty
		if token.RefreshToken != "initial-refresh-token" {
			t.Errorf("RefreshToken = %q", token.RefreshToken)
		}
	})

	// --- Step 4: Fetch playlist ---
	t.Run("fetch and format playlist", func(t *testing.T) {
		token, err := store.Load(ctx)
		if err != nil {
			t.Fatal(err)
		}

		client := spotify.Client{
			HTTPClient: spotifyServer.Client(),
			BaseURL:    spotifyServer.URL,
		}

		playlist, err := client.GetPlaylist(ctx, token.AccessToken, "test-playlist-123")
		if err != nil {
			t.Fatal(err)
		}

		if playlist.Name != "Test Playlist" {
			t.Errorf("Name = %q", playlist.Name)
		}
		if len(playlist.Tracks) != 3 {
			t.Fatalf("Tracks = %d, want 3", len(playlist.Tracks))
		}

		// --- Step 5: Format playlist ---
		output := spotify.FormatPlaylist(playlist)

		if !strings.Contains(output, "Test Playlist") {
			t.Error("output missing playlist name")
		}
		if !strings.Contains(output, "3 tracks") {
			t.Error("output missing track count")
		}
		if !strings.Contains(output, "12 min") {
			t.Error("output missing total duration")
		}
		if !strings.Contains(output, "now playing... Artist A - First Song") {
			t.Error("output missing first track")
		}
		if !strings.Contains(output, "now playing... Artist B, Artist C - Second Song") {
			t.Error("output missing second track with multiple artists")
		}
		if !strings.Contains(output, "now playing... Artist D - Third Song") {
			t.Error("output missing third track")
		}
	})

	// --- Step 6: Parse playlist URL ---
	t.Run("parse playlist URL extracts ID", func(t *testing.T) {
		id, err := spotify.ParsePlaylistID("https://open.spotify.com/playlist/test-playlist-123?si=abc")
		if err != nil {
			t.Fatal(err)
		}
		if id != "test-playlist-123" {
			t.Errorf("id = %q", id)
		}
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
