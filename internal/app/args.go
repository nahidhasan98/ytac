package app

import (
	"fmt"
	"strconv"
	"strings"
)

const defaultSearchCount = 3

func printHelp() {
	fmt.Println("🎵 YouTube Audio Player (ytac)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ytac [--verbose] [--search-count <1-10>] <url>          Play audio from YouTube URL")
	fmt.Println("  ytac [--verbose] [--search-count <1-10>] <search terms> Search and play")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -v, --verbose               Show full mpv logs during playback")
	fmt.Println("  -c, --search-count <1-10>   Number of search results to fetch (default: 3)")
	fmt.Println()
	fmt.Println("Keyboard Controls in Search (FZF):")
	fmt.Println("  ↑/↓ or j/k    Navigate up/down")
	fmt.Println("  Enter         Select and play")
	fmt.Println("  Ctrl+c / Esc  Cancel")
	fmt.Println("  /             Toggle search mode")
	fmt.Println()
	fmt.Println("Keyboard Controls in Player (MPV):")
	fmt.Println("  Space         Pause/Resume")
	fmt.Println("  →/←           Seek forward/backward 5s")
	fmt.Println("  ↑/↓           Volume up/down (5%)")
	fmt.Println("  > / <         Next/Previous track in search")
	fmt.Println("  q             Stop and return to menu")
	fmt.Println("  m             Mute/Unmute")
	fmt.Println("  9/0           Volume down/up (2%)")
	fmt.Println("  [/]           Speed down/up")
	fmt.Println("  Backspace     Reset speed to 1.0x")
	fmt.Println("  r             Restart from beginning")
	fmt.Println("  l             Loop entire track (toggle)")
	fmt.Println("  a             A-B loop (set start, end, clear)")
}

func parseArgs(args []string) (string, bool, int, error) {
	showVerbose := false
	searchCount := defaultSearchCount
	queryParts := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-v", "--verbose":
			showVerbose = true
		case "-c", "--search-count":
			i++
			if i >= len(args) {
				return "", showVerbose, 0, fmt.Errorf("missing value for %s", arg)
			}

			value, err := parseSearchCount(args[i])
			if err != nil {
				return "", showVerbose, 0, err
			}
			searchCount = value
		default:
			if strings.HasPrefix(arg, "--search-count=") {
				valueText := strings.TrimPrefix(arg, "--search-count=")
				value, err := parseSearchCount(valueText)
				if err != nil {
					return "", showVerbose, 0, err
				}
				searchCount = value
				continue
			}
			queryParts = append(queryParts, arg)
		}
	}

	query := strings.TrimSpace(strings.Join(queryParts, " "))
	if query == "" {
		return "", showVerbose, 0, fmt.Errorf("missing search terms or URL")
	}

	return query, showVerbose, searchCount, nil
}

func parseSearchCount(valueText string) (int, error) {
	value, err := strconv.Atoi(valueText)
	if err != nil {
		return 0, fmt.Errorf("invalid search count %q: must be a number between 1 and 10", valueText)
	}
	if value < 1 || value > 10 {
		return 0, fmt.Errorf("search count must be between 1 and 10")
	}
	return value, nil
}
