package spotify

import "testing"

func TestParsePlaylistID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "full URL",
			input: "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:  "URL with query params",
			input: "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M?si=abc123",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:  "bare playlist ID",
			input: "37i9dQZF1DXcBWIGoYBM5M",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:  "spotify URI",
			input: "spotify:playlist:37i9dQZF1DXcBWIGoYBM5M",
			want:  "37i9dQZF1DXcBWIGoYBM5M",
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid URL with wrong path",
			input:   "https://open.spotify.com/track/abc123",
			wantErr: true,
		},
		{
			name:    "invalid URI with wrong type",
			input:   "spotify:track:abc123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePlaylistID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
