package player

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
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
`
}

func extractURLsFromResults(results []youtube.Result) []string {
	urls := make([]string, 0, len(results))
	for _, result := range results {
		urls = append(urls, result.URL)
	}
	return urls
}

func buildMPVArgs(selectedIndex int, urls []string, showVerbose bool) []string {
	mpvArgs := []string{
		"--no-video",
		"--ytdl-format=bestaudio",
		"--term-osd-bar",
		"--term-playing-msg=▶️  Playing: 🎵 ${media-title}\n",
	}

	if !showVerbose {
		mpvArgs = append(mpvArgs, "--msg-level=all=no,term-msg=info,statusline=status")
	}

	mpvArgs = append(mpvArgs,
		"--input-default-bindings=no",
		"--input-builtin-bindings=yes",
		"--input-conf=/dev/stdin",
		fmt.Sprintf("--playlist-start=%d", selectedIndex),
	)
	mpvArgs = append(mpvArgs, urls...)

	return mpvArgs
}

func playMPV(mpvArgs []string) {
	play := exec.Command("mpv", mpvArgs...)
	play.Stdout = os.Stdout
	play.Stderr = os.Stdout
	play.Stdin = strings.NewReader(getKeyBindings())
	err := play.Run()
	if err == nil {
		return
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 0 {
		return
	}

	fmt.Printf("❌ Playback ended with an error: %v\n", err)
	fmt.Println("💡 Re-run with --verbose to show full mpv output for debugging.")
}

func PlaySearchResults(results []youtube.Result, selectedIndex int, showVerbose bool) {
	urls := extractURLsFromResults(results)
	fmt.Println("⌨️  Controls: Space=Pause | ←→=Seek | ↑↓=Volume | q=Stop&Menu | >/<=Next/Prev")
	fmt.Println()

	mpvArgs := buildMPVArgs(selectedIndex-1, urls, showVerbose)
	playMPV(mpvArgs)
}

func PlayDirectURL(url string, showVerbose bool) {
	fmt.Println("▶️  Playing from URL...")
	fmt.Println("⌨️  Controls: Space=Pause | ←→=Seek | ↑↓=Volume | q=Quit")

	mpvArgs := []string{
		"--no-video",
		"--ytdl-format=bestaudio",
		"--term-osd-bar",
		"--term-playing-msg=▶️  Playing: 🎵 ${media-title}\n",
		"--input-default-bindings=no",
		"--input-builtin-bindings=yes",
		"--input-conf=/dev/stdin",
	}
	if !showVerbose {
		mpvArgs = append(mpvArgs, "--msg-level=all=no,term-msg=info,statusline=status")
	}
	mpvArgs = append(mpvArgs, url)

	playMPV(mpvArgs)
}
