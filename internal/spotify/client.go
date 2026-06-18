package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var ErrForbidden = errors.New("you don't have access to this playlist — you must be the owner or a collaborator")

const (
	defaultBaseURL  = "https://api.spotify.com"
	defaultPageSize = 50
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	PageSize   int
}

type Track struct {
	Name       string
	Artists    []string
	Album      string
	DurationMs int
}

type Playlist struct {
	Name   string
	Tracks []Track
}

func (c *Client) GetPlaylist(ctx context.Context, token, playlistID string) (Playlist, error) {
	baseURL := c.baseURL()
	pageSize := c.pageSize()

	// First call: get playlist name + first page of tracks
	playlistURL := fmt.Sprintf("%s/v1/playlists/%s", baseURL, url.PathEscape(playlistID))
	var firstResp struct {
		Name   string `json:"name"`
		Tracks struct {
			Total int             `json:"total"`
			Items []trackItemJSON `json:"items"`
		} `json:"tracks"`
	}

	if err := c.doGet(ctx, token, playlistURL, &firstResp); err != nil {
		return Playlist{}, err
	}

	tracks := parseTracks(firstResp.Tracks.Items)
	total := firstResp.Tracks.Total

	// Paginate remaining tracks
	for offset := len(tracks); offset < total; offset += pageSize {
		pageURL := fmt.Sprintf("%s/v1/playlists/%s/tracks?offset=%s&limit=%s",
			baseURL, url.PathEscape(playlistID), strconv.Itoa(offset), strconv.Itoa(pageSize))

		var pageResp struct {
			Total int             `json:"total"`
			Items []trackItemJSON `json:"items"`
		}

		if err := c.doGet(ctx, token, pageURL, &pageResp); err != nil {
			return Playlist{}, err
		}

		tracks = append(tracks, parseTracks(pageResp.Items)...)
	}

	return Playlist{
		Name:   firstResp.Name,
		Tracks: tracks,
	}, nil
}

func (c *Client) doGet(ctx context.Context, token, requestURL string, dest any) error {
	return c.doGetWithRetry(ctx, token, requestURL, dest, 1)
}

func (c *Client) doGetWithRetry(ctx context.Context, token, requestURL string, dest any, retries int) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests && retries > 0 {
		retryAfter := resp.Header.Get("Retry-After")
		wait := 1 * time.Second
		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			wait = time.Duration(seconds) * time.Second
		}
		time.Sleep(wait)
		return c.doGetWithRetry(ctx, token, requestURL, dest, retries-1)
	}

	if resp.StatusCode == http.StatusForbidden {
		return ErrForbidden
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("spotify API error: status %d: %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *Client) baseURL() string {
	if c.BaseURL != "" {
		return c.BaseURL
	}
	return defaultBaseURL
}

func (c *Client) pageSize() int {
	if c.PageSize > 0 {
		return c.PageSize
	}
	return defaultPageSize
}

type trackItemJSON struct {
	Track struct {
		Name    string `json:"name"`
		Artists []struct {
			Name string `json:"name"`
		} `json:"artists"`
		Album struct {
			Name string `json:"name"`
		} `json:"album"`
		DurationMs int `json:"duration_ms"`
	} `json:"track"`
}

func parseTracks(items []trackItemJSON) []Track {
	tracks := make([]Track, 0, len(items))
	for _, item := range items {
		artists := make([]string, 0, len(item.Track.Artists))
		for _, a := range item.Track.Artists {
			artists = append(artists, a.Name)
		}
		tracks = append(tracks, Track{
			Name:       item.Track.Name,
			Artists:    artists,
			Album:      item.Track.Album.Name,
			DurationMs: item.Track.DurationMs,
		})
	}
	return tracks
}
