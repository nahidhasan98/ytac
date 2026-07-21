package ui

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/nahidhasan98/ytac/internal/youtube"
)

func createFzfCommand() *exec.Cmd {
	return exec.Command("fzf",
		"--prompt=🎵 Select a video ▶ ",
		"--header=↑↓:navigate | Enter:play | Ctrl-c:exit | ESC:exit",
		"--border=rounded",
		"--margin=1",
		"--padding=1",
		"--color=fg:#ffffff,bg:#000000,hl:#5fff87",
		"--color=fg+:#ffffff,bg+:#262626,hl+:#5fff87",
		"--color=info:#af87ff,prompt:#5fff87,pointer:#ff87d7",
		"--color=marker:#ff87d7,spinner:#ff87d7,header:#87afff",
		"--pointer=▶",
		"--marker=✓",
		"--height=~100%",
		"--reverse",
		"--bind=ctrl-j:down,ctrl-k:up",
		"--bind=alt-j:preview-down,alt-k:preview-up",
		"--tabstop=4",
	)
}

func formatResultsForFzf(results []youtube.Result) []string {
	numberedResults := make([]string, 0, len(results))
	for i, result := range results {
		displayText := fmt.Sprintf("[%d] %s | %s | %s", i+1, result.Title, result.Uploader, result.Duration)
		numberedResults = append(numberedResults, displayText)
	}
	return numberedResults
}

func SelectVideo(results []youtube.Result) (int, bool) {
	fzf := createFzfCommand()
	numberedResults := formatResultsForFzf(results)

	fzf.Stdin = strings.NewReader(strings.Join(numberedResults, "\n"))
	fzf.Stderr = os.Stderr

	fzfOut, err := fzf.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := fzf.Start(); err != nil {
		log.Fatal(err)
	}

	fzfScanner := bufio.NewScanner(fzfOut)
	if !fzfScanner.Scan() {
		return 0, false
	}

	line := fzfScanner.Text()
	fzf.Wait()

	var selectedIndex int
	fmt.Sscanf(line, "[%d]", &selectedIndex)

	return selectedIndex, true
}
