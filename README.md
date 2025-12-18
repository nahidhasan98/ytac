# ytac - YouTube Audio in Console

A terminal-based YouTube audio player with keyboard controls for searching and playing YouTube audio.

## Features

- 🔍 Search YouTube directly from terminal
- 🎵 Play audio from YouTube URLs or search results
- ⌨️ Full keyboard controls for navigation and playback
- 📦 Auto-install missing dependencies
- 🎨 Beautiful interactive search interface with fzf
- ⚡ Live search results display
- 🎛️ MPV player controls (volume, seek, speed)

## Requirements

- `mpv` - Media player
- `yt-dlp` - YouTube downloader
- `fzf` - Fuzzy finder

The tool will offer to install missing dependencies automatically on first run.

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/ytac.git
cd ytac

# Build
./build.sh

# Run
bin/ytac <search terms>
```

## Usage

### Search and play
```bash
bin/ytac lofi hip hop
bin/ytac adele hello
```

### Play from URL
```bash
bin/ytac https://www.youtube.com/watch?v=VIDEO_ID
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
| `9`/`0` | Volume down/up 2% |
| `[`/`]` | Speed down/up |
| `Backspace` | Reset speed to 1.0x |
| `m` | Mute/Unmute |
| `q` | Quit player |

## Build from Source

```bash
# Standard build
./build.sh

# Manual build
CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/ytac
```
