package app

import (
	"fmt"
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
		player.PlayDirectURL(query, showVerbose)
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

		player.PlaySearchResults(results, selectedIndex, showVerbose)
		fmt.Println("\n🔄 Returning to menu...")
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("👋 Goodbye!")
	return 0
}
