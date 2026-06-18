# spotify-playlist

A CLI tool that displays formatted Spotify playlist tracks. Authenticate once, then fetch and display any playlist you own or collaborate on.

## Install

### [Scoop](https://scoop.sh/) (Windows)

```powershell
scoop bucket add dh85 https://github.com/dh85/scoop-bucket
scoop install spotify-playlist
```

### [Homebrew](https://brew.sh/) (macOS/Linux)

```bash
brew install dh85/tap/spotify-playlist
```

## Usage

```bash
spotify-playlist <playlist-url>
```

### Examples

```bash
# Full Spotify URL
spotify-playlist "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M"

# Spotify URI
spotify-playlist spotify:playlist:37i9dQZF1DXcBWIGoYBM5M

# Just the playlist ID
spotify-playlist 37i9dQZF1DXcBWIGoYBM5M

# Plain text output (for piping/copying)
spotify-playlist --raw "https://open.spotify.com/playlist/37i9dQZF1DXcBWIGoYBM5M"
```

### First run

On first use, the tool opens your browser to authorize with Spotify. After you approve, you'll be redirected — paste that URL back into the terminal:

```
Opening browser for Spotify login...

If the browser doesn't open, visit:
https://accounts.spotify.com/authorize?client_id=...

Paste the URL you were redirected to: https://google.com/?code=...&state=...
```

### Subsequent runs

No login needed. Tokens are stored in your system keyring and refreshed automatically.

## Output

### Default

```
──────────────────────────────────────────────────────────────────────
 Pepsi Blues 2026.06.04
 📊 17 tracks ⏱️  1 hr, 1 min, 25 sec
──────────────────────────────────────────────────────────────────────

<color=#FACC15><sprite name="musical-notes_1F3B6"> now playing... Otis Rush - Working Man <sprite name="musical-notes_1F3B6">
<color=#FACC15><sprite name="musical-notes_1F3B6"> now playing... Bonnie Raitt - Love Me Like a Man <sprite name="musical-notes_1F3B6">
<color=#FACC15><sprite name="musical-notes_1F3B6"> now playing... Tommy Castro - Ninety-Nine and One Half <sprite name="musical-notes_1F3B6">
...
```

### Raw (`--raw`)

```
Otis Rush - Working Man
Bonnie Raitt - Love Me Like a Man - 2008 Remaster
Tommy Castro - Ninety-Nine and One Half
...
```

## Requirements

- The playlist must be owned by you or you must be a collaborator

## Development

Requires [mise](https://mise.jdx.dev/) for toolchain management.

```bash
# Install Go and tools
mise install

# Download dependencies
make setup

# Build binary
make build

# Run tests
make test

# Run linter
make lint

# Run all checks
make check

# Run the app
SPOTIFY_CLIENT_ID=your-id go run ./cmd/spotify/ "https://open.spotify.com/playlist/..."
```

## Releasing

Releases are automated via GitHub Actions and [GoReleaser](https://goreleaser.com/).

```bash
git tag v0.1.0
git push origin v0.1.0
```

This will:
- Build binaries for Linux, macOS, and Windows (amd64 + arm64)
- Create a GitHub Release with archives and checksums
- Update the Homebrew formula in `dh85/homebrew-tap`
- Update the Scoop manifest in `dh85/scoop-bucket`

## License

MIT

## Alternative install methods

### Go

```bash
go install github.com/dh85/spotify-playlist/cmd/spotify@latest
```

Requires `SPOTIFY_CLIENT_ID` environment variable when using `go install`.

### Binary

Download from [GitHub Releases](https://github.com/dh85/spotify-playlist/releases).
