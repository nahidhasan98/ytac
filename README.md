# ytac - YouTube Audio in Console

A terminal-based YouTube audio player with keyboard controls for searching and playing YouTube audio.

## Features

- 🔍 Search YouTube directly from terminal
- 🎵 Play audio from YouTube URLs or search results
- ⌨️ Full keyboard controls for navigation and playback
- 📦 Auto-install missing dependencies (mpv, yt-dlp, fzf)
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
bin/ytac nasheed
bin/ytac tafseer
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
CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/ytac
```

## Tips

- All search results become a playlist - navigate with `>` and `<`
- Press `q` to return to menu anytime during playback
- Use `l` to loop your favorite track continuously
- Use `a` key to loop specific sections (A-B loop)
- Adjust search result count in code for more/fewer options
