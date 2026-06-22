package spotify

import (
	"fmt"
	"strings"
)

const (
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
	colorReset  = "\033[0m"
	boxWidth    = 70
	trackFormat = "<color=#FACC15><sprite name=\"musical-notes_1F3B6\"> now playing... %s - %s <sprite name=\"musical-notes_1F3B6\">\n"
)

func FormatPlaylist(p Playlist) string {
	if len(p.Tracks) == 0 {
		return ""
	}

	var b strings.Builder
	line := strings.Repeat("─", boxWidth)

	// Header
	fmt.Fprintf(&b, "\n%s%s%s\n", colorCyan, line, colorReset)
	fmt.Fprintf(&b, "%s%s %s%s%s%s\n", colorCyan, colorReset, colorBold, colorGreen, p.Name, colorReset)

	totalMs := 0
	for _, t := range p.Tracks {
		totalMs += t.DurationMs
	}
	duration := formatTotalDuration(totalMs)
	fmt.Fprintf(&b, "%s%s %s📊 %d tracks%s %s⏱️  %s%s\n",
		colorCyan, colorReset, colorYellow, len(p.Tracks), colorReset, colorYellow, duration, colorReset)

	fmt.Fprintf(&b, "%s%s%s\n\n", colorCyan, line, colorReset)

	// Track list
	for _, t := range p.Tracks {
		artists := strings.Join(t.Artists, ", ")
		fmt.Fprintf(&b, trackFormat, artists, t.Name)
	}

	return b.String() + "\n"
}

func FormatPlaylistCustom(p Playlist, style string) string {
	if len(p.Tracks) == 0 {
		return ""
	}

	var b strings.Builder
	for _, t := range p.Tracks {
		artists := strings.Join(t.Artists, ", ")
		line := strings.NewReplacer(
			"{artist}", artists,
			"{title}", t.Name,
			"{album}", t.Album,
		).Replace(style)
		fmt.Fprintln(&b, line)
	}
	return b.String()
}

func FormatPlaylistRaw(p Playlist) string {
	if len(p.Tracks) == 0 {
		return ""
	}

	var b strings.Builder
	for _, t := range p.Tracks {
		artists := strings.Join(t.Artists, ", ")
		fmt.Fprintf(&b, "%s - %s\n", artists, t.Name)
	}
	return b.String()
}

func formatTotalDuration(ms int) string {
	totalSeconds := ms / 1000
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	var parts []string
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%d hr", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%d min", minutes))
	}
	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%d sec", seconds))
	}

	return strings.Join(parts, ", ")
}
