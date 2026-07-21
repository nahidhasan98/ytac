# ytac - YouTube Audio in Console

A terminal-based YouTube audio player with keyboard controls for searching and playing YouTube audio.

## Features

- 🔍 Search YouTube directly from terminal
- 🎵 Play audio from YouTube URLs or search results
- ⌨️ Full keyboard controls for navigation and playback
- 📦 Auto-install missing dependencies (mpv, yt-dlp, fzf)
- ⬆️ Auto-check yt-dlp against the latest GitHub release and update when needed
- 🎨 Beautiful interactive search interface with fzf
- ⚡ Live search results display as they arrive
- 🎛️ Complete MPV player controls (volume, seek, speed, loop)
- 🔄 Playlist navigation (next/previous through search results)
- ↩️ Return to menu after playback ends
- 🔁 Loop controls (entire track or A-B section)
- 🎯 Clean, clutter-free output interface
- 📱 User-friendly notifications for track changes

## Requirements

- `mpv` - Media player
- `yt-dlp` - YouTube downloader
- `fzf` - Fuzzy finder

The tool will offer to install missing dependencies automatically on first run.
For `yt-dlp`, it checks the latest stable release from `https://github.com/yt-dlp/yt-dlp` and installs or updates the appropriate binary according to the OS when possible.

## Installation

```bash
# Clone the repository
git clone https://github.com/nahidhasan98/ytac.git
cd ytac

# Build
./build.sh

# Run
bin/ytac <search terms>
```

## Usage

### Search and play
```bash
bin/ytac nasheed
bin/ytac tafseer
```

### Play from URL
```bash
bin/ytac https://www.youtube.com/watch?v=VIDEO_ID
```

### Verbose playback logs
```bash
bin/ytac --verbose nasheed
bin/ytac -v https://www.youtube.com/watch?v=VIDEO_ID
```

### Control search result count
```bash
bin/ytac --search-count 3 nasheed
bin/ytac -c 10 tafseer
```

### Show help
```bash
bin/ytac --help
```

## Keyboard Controls

### Search Navigation (FZF)
| Key | Action |
|-----|--------|
| `↑`/`↓` or `j`/`k` | Navigate up/down |
| `Enter` | Select and play |
| `Ctrl+c` / `Esc` | Cancel |
| `/` | Toggle search mode |

### Playback Controls (MPV)
| Key | Action |
|-----|--------|
| `Space` | Pause/Resume |
| `→`/`←` | Seek forward/backward 5s |
| `↑`/`↓` | Volume up/down 5% |
| `>`/`<` | Next/Previous track in playlist |
| `9`/`0` | Volume down/up 2% |
| `[`/`]` | Speed down/up |
| `Backspace` | Reset speed to 1.0x |
| `r` | Restart from beginning |
| `l` | Loop entire track (toggle on/off) |
| `a` | A-B loop (set start, end, clear) |
| `m` | Mute/Unmute |
| `q` | Stop and return to menu |

## How It Works

1. **Search**: Enter your search terms to find YouTube videos
2. **Select**: Browse results with arrow keys and press Enter
3. **Play**: Audio starts playing with all search results as a playlist
4. **Navigate**: Use `>` and `<` to move between tracks
5. **Return**: Press `q` to stop and return to the selection menu
6. **Repeat**: Select another video or exit

## Build from Source

```bash
# Standard build
./build.sh

# Manual build
CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/ytac ./cmd/ytac
```

## Development

```bash
go test ./...
go vet ./...
```

## OS Notes

- `yt-dlp` installation is OS-aware and selects a release asset based on the platform
- The overall app is best supported on Linux and macOS right now because dependency installation still assumes `apt`, `dnf`, `pacman`, or `brew`, and playback uses `/dev/stdin` for mpv bindings

## Tips

- All search results become a playlist - navigate with `>` and `<`
- Press `q` to return to menu anytime during playback
- Use `l` to loop your favorite track continuously
- Use `a` key to loop specific sections (A-B loop)
- Use `--search-count` or `-c` to choose how many results to fetch, from 1 to 10
- Add `--verbose` when you want full mpv logs for troubleshooting playback issues

## License

MIT
