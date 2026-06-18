package spotify

import (
	"fmt"
	"net/url"
	"strings"
)

func ParsePlaylistID(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty playlist input")
	}

	// spotify:playlist:ID
	if strings.HasPrefix(input, "spotify:") {
		parts := strings.Split(input, ":")
		if len(parts) == 3 && parts[1] == "playlist" && parts[2] != "" {
			return parts[2], nil
		}
		return "", fmt.Errorf("invalid spotify URI: %s", input)
	}

	// https://open.spotify.com/playlist/ID
	if strings.Contains(input, "/") {
		parsed, err := url.Parse(input)
		if err != nil {
			return "", err
		}
		parts := strings.Split(strings.TrimPrefix(parsed.Path, "/"), "/")
		if len(parts) == 2 && parts[0] == "playlist" && parts[1] != "" {
			return parts[1], nil
		}
		return "", fmt.Errorf("invalid spotify URL: %s", input)
	}

	// Bare playlist ID
	return input, nil
}
