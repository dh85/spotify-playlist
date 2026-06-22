package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const configFile = ".spotify-playlist"

type Config struct {
	OutputDir   string
	SaveToFile  bool
	FormatStyle string
}

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configFile), nil
}

func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{OutputDir: "."}, err
	}
	return LoadFrom(path)
}

func LoadFrom(path string) (Config, error) {
	cfg := Config{OutputDir: ".", SaveToFile: true}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "output_dir":
			cfg.OutputDir = strings.TrimSpace(value)
		case "save_to_file":
			cfg.SaveToFile = strings.TrimSpace(value) == "true"
		case "format_style":
			cfg.FormatStyle = strings.TrimSpace(value)
		}
	}
	return cfg, scanner.Err()
}

func Init() error {
	path, err := Path()
	if err != nil {
		return err
	}
	return InitAt(path)
}

func InitAt(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	content := `# spotify-playlist config
# Save playlist to a .txt file and open it (true/false)
save_to_file=true
# Directory where playlist .txt files are saved
output_dir=.
# Custom track format (available: {artist}, {title}, {album})
# format_style={artist} - {title}
`
	fmt.Printf("Created config: %s\n", path)
	return os.WriteFile(path, []byte(content), 0o644)
}
