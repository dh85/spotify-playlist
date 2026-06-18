package main

import (
	"os"
	"testing"
)

func TestRun(t *testing.T) {
	t.Run("no args prints help", func(t *testing.T) {
		os.Args = []string{"spotify-playlist"}
		err := run()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("--help flag", func(t *testing.T) {
		os.Args = []string{"spotify-playlist", "--help"}
		err := run()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("-h flag", func(t *testing.T) {
		os.Args = []string{"spotify-playlist", "-h"}
		err := run()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("--version flag", func(t *testing.T) {
		os.Args = []string{"spotify-playlist", "--version"}
		err := run()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("-v flag", func(t *testing.T) {
		os.Args = []string{"spotify-playlist", "-v"}
		err := run()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("missing client ID returns error", func(t *testing.T) {
		os.Args = []string{"spotify-playlist", "37i9dQZF1DXcBWIGoYBM5M"}
		t.Setenv("SPOTIFY_CLIENT_ID", "")
		original := defaultClientID
		defaultClientID = ""
		t.Cleanup(func() { defaultClientID = original })

		err := run()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "SPOTIFY_CLIENT_ID environment variable is required" {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("invalid playlist URL returns error", func(t *testing.T) {
		os.Args = []string{"spotify-playlist", "https://open.spotify.com/track/abc123"}
		t.Setenv("SPOTIFY_CLIENT_ID", "test-id")

		err := run()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("--raw flag is parsed alongside playlist URL", func(t *testing.T) {
		os.Args = []string{"spotify-playlist", "--raw", "37i9dQZF1DXcBWIGoYBM5M"}
		t.Setenv("SPOTIFY_CLIENT_ID", "")
		original := defaultClientID
		defaultClientID = ""
		t.Cleanup(func() { defaultClientID = original })

		err := run()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		// Proves --raw was consumed and playlist ID was parsed correctly
		if err.Error() != "SPOTIFY_CLIENT_ID environment variable is required" {
			t.Errorf("err = %v", err)
		}
	})

	t.Run("-r flag is parsed alongside playlist URL", func(t *testing.T) {
		os.Args = []string{"spotify-playlist", "-r", "37i9dQZF1DXcBWIGoYBM5M"}
		t.Setenv("SPOTIFY_CLIENT_ID", "")
		original := defaultClientID
		defaultClientID = ""
		t.Cleanup(func() { defaultClientID = original })

		err := run()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "SPOTIFY_CLIENT_ID environment variable is required" {
			t.Errorf("err = %v", err)
		}
	})
}

func TestResolveClientID(t *testing.T) {
	t.Run("env var takes precedence over default", func(t *testing.T) {
		t.Setenv("SPOTIFY_CLIENT_ID", "env-id")
		original := defaultClientID
		defaultClientID = "default-id"
		t.Cleanup(func() { defaultClientID = original })

		got := resolveClientID()
		if got != "env-id" {
			t.Errorf("got %q, want %q", got, "env-id")
		}
	})

	t.Run("falls back to defaultClientID when env var is empty", func(t *testing.T) {
		t.Setenv("SPOTIFY_CLIENT_ID", "")
		original := defaultClientID
		defaultClientID = "baked-in"
		t.Cleanup(func() { defaultClientID = original })

		got := resolveClientID()
		if got != "baked-in" {
			t.Errorf("got %q, want %q", got, "baked-in")
		}
	})

	t.Run("returns empty when neither is set", func(t *testing.T) {
		t.Setenv("SPOTIFY_CLIENT_ID", "")
		original := defaultClientID
		defaultClientID = ""
		t.Cleanup(func() { defaultClientID = original })

		got := resolveClientID()
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}
