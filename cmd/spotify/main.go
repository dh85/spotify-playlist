package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/dh85/spotify-playlist/internal/auth"
	"github.com/dh85/spotify-playlist/internal/config"
	"github.com/dh85/spotify-playlist/internal/spotify"
	"github.com/dh85/spotify-playlist/internal/storage"
)

// Set at build time via:
//
//	go build -ldflags "-X main.defaultClientID=your-id -X main.version=v0.1.0" ./cmd/spotify/
var (
	defaultClientID string
	version         = "dev"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	raw := false
	args := os.Args[1:]

	var filtered []string
	for _, arg := range args {
		switch arg {
		case "--help", "-h":
			printHelp()
			return nil
		case "--version", "-v":
			fmt.Println(version)
			return nil
		case "--raw", "-r":
			raw = true
		default:
			filtered = append(filtered, arg)
		}
	}

	if len(filtered) < 1 {
		printHelp()
		return nil
	}

	playlistID, err := spotify.ParsePlaylistID(filtered[0])
	if err != nil {
		return err
	}

	ctx := context.Background()

	clientID := resolveClientID()
	if clientID == "" {
		return fmt.Errorf("SPOTIFY_CLIENT_ID environment variable is required")
	}

	tokenStore := storage.KeyringTokenStore{
		Service: "spotify-playlist-cli",
		User:    "default",
	}

	oauthClient := auth.OAuthClient{
		ClientID:    clientID,
		RedirectURI: "https://google.com/",
		Scopes:      []string{"playlist-read-private", "playlist-read-collaborative"},
	}

	refresher := auth.SpotifyTokenRefresher{
		ClientID: clientID,
	}

	tokenManager := auth.NewTokenManager(&tokenStore, &refresher, time.Now)

	token, err := tokenManager.ValidToken(ctx)
	if errors.Is(err, auth.ErrLoginRequired) {
		loginCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		flow := auth.LoginFlow{
			OAuthClient:    oauthClient,
			Store:          &tokenStore,
			BrowserOpener:  openBrowser,
			RedirectReader: readRedirectURL,
		}

		fmt.Println("Opening browser for Spotify login...")
		if err := flow.Run(loginCtx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("login timed out — please try again")
			}
			return fmt.Errorf("login failed: %w", err)
		}

		token, err = tokenManager.ValidToken(ctx)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	client := spotify.Client{HTTPClient: &http.Client{Timeout: 10 * time.Second}}

	playlist, err := client.GetPlaylist(ctx, token.AccessToken, playlistID)
	if err != nil {
		return err
	}

	if raw {
		fmt.Print(spotify.FormatPlaylistRaw(playlist))
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := config.Init(); err != nil {
		return err
	}

	if cfg.SaveToFile {
		return saveAndOpen(playlist, cfg)
	}

	if cfg.FormatStyle != "" {
		fmt.Print(spotify.FormatPlaylistCustom(playlist, cfg.FormatStyle))
	} else {
		fmt.Print(spotify.FormatPlaylist(playlist))
	}
	return nil
}

func printHelp() {
	fmt.Print(`spotify-playlist - display formatted Spotify playlist tracks

Usage:
  spotify-playlist [flags] <playlist-url>

Flags:
  -r, --raw        Plain text output to stdout (skips file save)
  -v, --version    Print version
  -h, --help       Print this help

By default, output is saved to a .txt file and opened.
Config: ~/.config/spotify-playlist/config

Examples:
  spotify-playlist "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M"
  spotify-playlist --raw spotify:playlist:37i9dQZF1DXcBWIGoYBM5M
  spotify-playlist 37i9dQZF1DXcBWIGoYBM5M
`)
}

func openBrowser(u string) error {
	fmt.Printf("\nIf the browser doesn't open, visit:\n%s\n", u)
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("wslview"); err == nil {
			return exec.Command("wslview", u).Start()
		}
		return exec.Command("xdg-open", u).Start()
	case "darwin":
		return exec.Command("open", u).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

var sanitizeRe = regexp.MustCompile(`[^a-zA-Z0-9_\-. ]+`)

func saveAndOpen(p spotify.Playlist, cfg config.Config) error {
	name := sanitizeRe.ReplaceAllString(p.Name, "")
	filename := filepath.Join(cfg.OutputDir, name+".txt")

	var content string
	if cfg.FormatStyle != "" {
		content = spotify.FormatPlaylistCustom(p, cfg.FormatStyle)
	} else {
		content = spotify.FormatPlaylistRaw(p)
	}
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		return err
	}

	fmt.Printf("Saved to %s\n", filename)
	return openFile(filename)
}

func openFile(path string) error {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("wslview"); err == nil {
			return exec.Command("wslview", path).Start()
		}
		return exec.Command("xdg-open", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", path).Start()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func readRedirectURL() (string, error) {
	fmt.Print("\nPaste the URL you were redirected to: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func resolveClientID() string {
	if id := os.Getenv("SPOTIFY_CLIENT_ID"); id != "" {
		return id
	}
	return defaultClientID
}
