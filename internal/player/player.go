package player

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nahidhasan98/ytac/internal/youtube"
)

func getKeyBindings() string {
	return `# Custom key bindings for ytac
SPACE cycle pause
RIGHT seek 5
LEFT seek -5
UP add volume 5
DOWN add volume -5
q quit
m cycle mute
9 add volume -2
0 add volume 2
> playlist-next ; show-text "⏭️  Next Track" 4000
< playlist-prev ; show-text "⏮️  Previous Track" 4000
[ multiply speed 0.9091
] multiply speed 1.1
BS set speed 1.0
r seek 0 absolute
l cycle-values loop-file "inf" "no"
a ab-loop
c cycle sub
`
}

//go:embed caption.lua
var captionScript []byte

// statusMsg renders the live caption (published by caption.lua) above the
// playback position. mpv redraws the multi-line status block in place.
const statusMsg = "${user-data/ytac/caption}\n" +
	"${?pause==yes:(Paused) }A: ${time-pos} / ${duration} (${percent-pos}%)  Vol: ${volume}%"

// terminalWidth returns the terminal's column count, or 80 if unknown.
func terminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return 80
	}
	var rows, cols int
	if _, err := fmt.Sscan(string(out), &rows, &cols); err != nil || cols <= 0 {
		return 80
	}
	return cols
}

// writeCaptionScript writes the embedded Lua helper to a fixed path in the
// user cache dir (overwritten each run, so nothing accumulates).
func writeCaptionScript() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	dir = filepath.Join(dir, "ytac")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "caption.lua")
	if err := os.WriteFile(path, captionScript, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func baseMPVArgs(showVerbose bool) []string {
	args := []string{
		"--no-video",
		"--ytdl-format=bestaudio",
		fmt.Sprintf("--script-opts=ytdl_hook-ytdl_path=yt-dlp,ytac-width=%d", terminalWidth()),
		// mpv prints subtitles to the terminal itself when there is no video;
		// hide that so only our single-line caption is shown.
		"--sub-visibility=no",
		// Fetch English captions (manual first, auto-generated as fallback)
		// so they can be shown in the terminal status line.
		"--ytdl-raw-options-append=write-subs=",
		"--ytdl-raw-options-append=write-auto-subs=",
		"--ytdl-raw-options-append=sub-langs=en.*",
		"--slang=en",
		"--term-osd-bar",
		"--term-playing-msg=▶️  Playing: 🎵 ${media-title}\n",
		"--term-status-msg=" + statusMsg,
	}
	if !showVerbose {
		args = append(args, "--msg-level=all=no,term-msg=info,statusline=status")
	}
	return append(args,
		"--input-default-bindings=no",
		"--input-builtin-bindings=yes",
		"--input-conf=/dev/stdin",
	)
}

func extractURLsFromResults(results []youtube.Result) []string {
	urls := make([]string, 0, len(results))
	for _, result := range results {
		urls = append(urls, result.URL)
	}
	return urls
}

func buildMPVArgs(selectedIndex int, urls []string, showVerbose bool) []string {
	mpvArgs := append(baseMPVArgs(showVerbose), fmt.Sprintf("--playlist-start=%d", selectedIndex))
	return append(mpvArgs, urls...)
}

func playMPV(mpvArgs []string) error {
	if script, err := writeCaptionScript(); err == nil {
		mpvArgs = append([]string{"--script=" + script}, mpvArgs...)
	}

	play := exec.Command("mpv", mpvArgs...)
	play.Stdout = os.Stdout
	play.Stderr = os.Stdout
	play.Stdin = strings.NewReader(getKeyBindings())
	err := play.Run()
	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 0 {
		return nil
	}

	return err
}

func PlaySearchResults(results []youtube.Result, selectedIndex int, showVerbose bool) error {
	urls := extractURLsFromResults(results)
	fmt.Println("⌨️  Controls: Space=Pause | ←→=Seek | ↑↓=Volume | q=Stop&Menu | >/<=Next/Prev | c=Captions")
	fmt.Println()

	mpvArgs := buildMPVArgs(selectedIndex-1, urls, showVerbose)
	return playMPV(mpvArgs)
}

func PlayDirectURL(url string, showVerbose bool) error {
	fmt.Println("▶️  Playing from URL...")
	fmt.Println("⌨️  Controls: Space=Pause | ←→=Seek | ↑↓=Volume | q=Quit | c=Captions")

	mpvArgs := append(baseMPVArgs(showVerbose), url)
	return playMPV(mpvArgs)
}
