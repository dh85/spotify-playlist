package spotify

import (
	"strings"
	"testing"
)

func TestFormatPlaylist(t *testing.T) {
	t.Run("formats playlist with header and tracks", func(t *testing.T) {
		playlist := Playlist{
			Name: "My Favorites",
			Tracks: []Track{
				{Name: "Bohemian Rhapsody", Artists: []string{"Queen"}, Album: "A Night at the Opera", DurationMs: 354000},
				{Name: "Get Lucky", Artists: []string{"Daft Punk", "Pharrell Williams"}, Album: "Random Access Memories", DurationMs: 248000},
			},
		}

		got := FormatPlaylist(playlist)

		assertContains(t, got, "My Favorites")
		assertContains(t, got, "2 tracks")
		assertContains(t, got, "10 min, 2 sec")
		assertContains(t, got, `<color=#FACC15><sprite name="musical-notes_1F3B6"> now playing... Queen - Bohemian Rhapsody <sprite name="musical-notes_1F3B6">`)
		assertContains(t, got, `<color=#FACC15><sprite name="musical-notes_1F3B6"> now playing... Daft Punk, Pharrell Williams - Get Lucky <sprite name="musical-notes_1F3B6">`)
	})

	t.Run("returns empty string for no tracks", func(t *testing.T) {
		playlist := Playlist{Name: "Empty", Tracks: nil}
		got := FormatPlaylist(playlist)
		if got != "" {
			t.Errorf("got: %q, want empty", got)
		}
	})

	t.Run("formats duration with hours", func(t *testing.T) {
		playlist := Playlist{
			Name: "Long Playlist",
			Tracks: []Track{
				{Name: "Song", Artists: []string{"A"}, DurationMs: 7261000},
			},
		}

		got := FormatPlaylist(playlist)
		assertContains(t, got, "2 hr, 1 min, 1 sec")
	})

	t.Run("formats all tracks for large playlists", func(t *testing.T) {
		tracks := make([]Track, 100)
		for i := range tracks {
			tracks[i] = Track{Name: "Song", Artists: []string{"Artist"}, DurationMs: 60000}
		}

		playlist := Playlist{Name: "Big", Tracks: tracks}
		got := FormatPlaylist(playlist)

		trackLines := strings.Count(got, "now playing...")
		if trackLines != 100 {
			t.Errorf("track lines = %d, want 100", trackLines)
		}
	})

	t.Run("ends with trailing newline for spacing", func(t *testing.T) {
		playlist := Playlist{
			Name:   "Test",
			Tracks: []Track{{Name: "S", Artists: []string{"A"}, DurationMs: 1000}},
		}

		got := FormatPlaylist(playlist)
		if !strings.HasSuffix(got, "\n\n") {
			t.Error("expected trailing blank line")
		}
	})
}

func TestFormatPlaylistRaw(t *testing.T) {
	t.Run("formats tracks as plain text", func(t *testing.T) {
		playlist := Playlist{
			Name: "My Playlist",
			Tracks: []Track{
				{Name: "Song A", Artists: []string{"Artist 1"}, DurationMs: 180000},
				{Name: "Song B", Artists: []string{"Artist 2", "Artist 3"}, DurationMs: 240000},
			},
		}

		got := FormatPlaylistRaw(playlist)
		want := "Artist 1 - Song A\nArtist 2, Artist 3 - Song B\n"

		if got != want {
			t.Errorf("got:\n%s\nwant:\n%s", got, want)
		}
	})

	t.Run("returns empty string for no tracks", func(t *testing.T) {
		got := FormatPlaylistRaw(Playlist{Name: "Empty"})
		if got != "" {
			t.Errorf("got: %q, want empty", got)
		}
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		ms   int
		want string
	}{
		{0, "0 sec"},
		{5000, "5 sec"},
		{60000, "1 min"},
		{62000, "1 min, 2 sec"},
		{3600000, "1 hr"},
		{3661000, "1 hr, 1 min, 1 sec"},
		{7200000, "2 hr"},
	}

	for _, tt := range tests {
		got := formatTotalDuration(tt.ms)
		if got != tt.want {
			t.Errorf("formatTotalDuration(%d) = %q, want %q", tt.ms, got, tt.want)
		}
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("output missing %q", substr)
	}
}
