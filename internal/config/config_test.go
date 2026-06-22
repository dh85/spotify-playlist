package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFrom(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantDir  string
		wantSave bool
	}{
		{"returns defaults when file not exists", "", ".", true},
		{"parses valid config", "output_dir=/tmp/playlists\nsave_to_file=true\n", "/tmp/playlists", true},
		{"save_to_file false", "save_to_file=false\n", ".", false},
		{"ignores comments and blank lines", "# comment\n  \noutput_dir=~/Music\n\n# another\ninvalid_line_without_equals\n", "~/Music", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config")
			if tt.content != "" {
				os.WriteFile(path, []byte(tt.content), 0o644)
			} else {
				path = "/nonexistent/path"
			}

			cfg, err := LoadFrom(path)
			if err != nil {
				t.Fatal(err)
			}
			if cfg.OutputDir != tt.wantDir {
				t.Errorf("got OutputDir=%q, want %q", cfg.OutputDir, tt.wantDir)
			}
			if cfg.SaveToFile != tt.wantSave {
				t.Errorf("got SaveToFile=%v, want %v", cfg.SaveToFile, tt.wantSave)
			}
		})
	}
}

func TestInitAt_CreatesFileWhenNotExists(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "config")

	if err := InitAt(tmp); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OutputDir != "." {
		t.Errorf("got OutputDir=%q, want %q", cfg.OutputDir, ".")
	}
}

func TestInitAt_NoOpWhenFileExists(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "config")
	original := "output_dir=/custom\n"
	os.WriteFile(tmp, []byte(original), 0o644)

	if err := InitAt(tmp); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(tmp)
	if string(data) != original {
		t.Error("InitAt overwrote existing config file")
	}
}
