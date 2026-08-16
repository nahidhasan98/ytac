package app

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/nahidhasan98/ytac/internal/deps"
	"github.com/nahidhasan98/ytac/internal/player"
	"github.com/nahidhasan98/ytac/internal/ui"
	"github.com/nahidhasan98/ytac/internal/youtube"
)

func Run(args []string) int {
	if len(args) < 1 {
		printHelp()
		return 1
	}

	if args[0] == "-h" || args[0] == "--help" {
		printHelp()
		return 0
	}

	deps.Ensure()

	query, showVerbose, searchCount, err := parseArgs(args)
	if err != nil {
		fmt.Printf("❌ %v\n\n", err)
		printHelp()
		return 1
	}

	if strings.HasPrefix(query, "http") {
		if err := player.PlayDirectURL(query, showVerbose); err != nil {
			handlePlaybackError(err, []string{query}, showVerbose, func() error {
				return player.PlayDirectURL(query, showVerbose)
			})
		}
		return 0
	}

	fmt.Printf("🔍 Searching YouTube for: \"%s\"\n\n", query)
	results, err := youtube.Search(query, searchCount)
	if err != nil {
		fmt.Printf("❌ Search failed: %v\n", err)
		return 1
	}

	if len(results) == 0 {
		fmt.Println("❌ No results found")
		return 0
	}

	fmt.Printf("✅ Found %d results\n", len(results))

	for {
		selectedIndex, ok := ui.SelectVideo(results)
		if !ok {
			fmt.Println("\n👋 Exiting")
			break
		}

		if err := player.PlaySearchResults(results, selectedIndex, showVerbose); err != nil {
			urls := playableURLs(results)
			handlePlaybackError(err, urls, showVerbose, func() error {
				return player.PlaySearchResults(results, selectedIndex, showVerbose)
			})
		}
		fmt.Println("\n🔄 Returning to menu...")
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("👋 Goodbye!")
	return 0
}

func handlePlaybackError(err error, urls []string, showVerbose bool, retry func() error) {
	fmt.Printf("❌ Playback ended with an error: %v\n", err)
	if !showVerbose {
		fmt.Println("💡 Re-run with --verbose to show full mpv output for debugging.")
	}

	if !containsYouTubeURL(urls) || !deps.PromptUpdateYTDLPToNightly() {
		return
	}

	fmt.Println("🔁 Retrying playback with the updated yt-dlp...")
	if retryErr := retry(); retryErr != nil {
		fmt.Printf("❌ Playback still failed after updating yt-dlp: %v\n", retryErr)
	}
}

func playableURLs(results []youtube.Result) []string {
	urls := make([]string, 0, len(results))
	for _, result := range results {
		urls = append(urls, result.URL)
	}
	return urls
}

func containsYouTubeURL(urls []string) bool {
	for _, rawURL := range urls {
		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			continue
		}

		host := strings.ToLower(parsedURL.Hostname())
		if host == "youtube.com" || host == "youtu.be" || strings.HasSuffix(host, ".youtube.com") {
			return true
		}
	}

	return false
}
