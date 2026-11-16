# Changelog

All notable changes to ytac are listed here, newest first.
The newest `## [x.y.z]` heading is the version ytac reports, so add a new
entry at the top for every release.

## [1.0.0] - 2026-11-16

First versioned release.

### Added
- Live captions under the "Playing" line (English, auto-generated as fallback); `c` toggles them.
- Version shown on the first line of every run; `-V`/`--version` also prints this changelog entry.

### Fixed
- URLs pasted in quotes in zsh (`watch\?v\=...`) no longer fail to play.
- YouTube auto-captions no longer repeat the previous line.

### Earlier (before versioning)
- Search YouTube and pick a result with fzf; results play as a playlist (`>`/`<` to skip).
- Play directly from a YouTube URL.
- `--search-count` and `--verbose` options.
- yt-dlp update check, with a nightly-update retry when YouTube playback fails.
