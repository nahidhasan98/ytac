package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func checkDependency(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func installPackage(pkg string) error {
	fmt.Printf("📦 Installing %s...\n", pkg)

	// Detect package manager and install
	var cmd *exec.Cmd

	if checkDependency("apt-get") {
		cmd = exec.Command("sudo", "apt-get", "install", "-y", pkg)
	} else if checkDependency("dnf") {
		cmd = exec.Command("sudo", "dnf", "install", "-y", pkg)
	} else if checkDependency("pacman") {
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", pkg)
	} else if checkDependency("brew") {
		cmd = exec.Command("brew", "install", pkg)
	} else {
		return fmt.Errorf("no supported package manager found (apt, dnf, pacman, brew)")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ensureDependencies() {
	deps := map[string]string{
		"mpv":    "mpv",
		"yt-dlp": "yt-dlp",
		"fzf":    "fzf",
	}

	missing := []string{}
	for cmd := range deps {
		if !checkDependency(cmd) {
			missing = append(missing, cmd)
		}
	}

	if len(missing) == 0 {
		return
	}

	fmt.Println("⚠️  Missing required dependencies:")
	for _, dep := range missing {
		fmt.Printf("   - %s\n", dep)
	}
	fmt.Println()
	fmt.Print("📥 Would you like to install them now? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	if response != "y" && response != "yes" {
		fmt.Println("❌ Cannot proceed without required dependencies")
		os.Exit(1)
	}

	fmt.Println()
	for _, dep := range missing {
		if err := installPackage(deps[dep]); err != nil {
			log.Fatalf("❌ Failed to install %s: %v\n", dep, err)
		}
		fmt.Printf("✅ %s installed successfully\n", dep)
	}
	fmt.Println()
}

func run(cmd string, args ...string) *exec.Cmd {
	c := exec.Command(cmd, args...)
	c.Stderr = os.Stderr
	return c
}

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

func createFzfCommand() *exec.Cmd {
	return run("fzf",
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

func formatResultsForFzf(results []string) []string {
	numberedResults := []string{}
	for i, result := range results {
		parts := strings.Split(result, " | ")
		if len(parts) >= 4 {
			displayText := fmt.Sprintf("[%d] %s | %s | %s", i+1, parts[0], parts[1], parts[2])
			numberedResults = append(numberedResults, displayText)
		}
	}
	return numberedResults
}

func extractURLsFromResults(results []string) []string {
	urls := []string{}
	for _, result := range results {
		parts := strings.Split(result, " | ")
		if len(parts) >= 4 {
			urls = append(urls, parts[len(parts)-1])
		}
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

func getUserSelection(results []string) (int, bool) {
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

func playMPV(mpvArgs []string) {
	play := exec.Command("mpv", mpvArgs...)
	play.Stdout = os.Stdout
	play.Stderr = os.Stdout
	play.Stdin = strings.NewReader(getKeyBindings())
	play.Run() // Ignore errors, they're normal for quit
}

func selectAndPlay(results []string, showVerbose bool) bool {
	selectedIndex, ok := getUserSelection(results)
	if !ok {
		fmt.Println("\n👋 Exiting")
		return false
	}

	urls := extractURLsFromResults(results)
	fmt.Println("⌨️  Controls: Space=Pause | ←→=Seek | ↑↓=Volume | q=Stop&Menu | >/<=Next/Prev")
	fmt.Println()

	mpvArgs := buildMPVArgs(selectedIndex-1, urls, showVerbose)
	playMPV(mpvArgs)

	return true
}

func printHelp() {
	fmt.Println("🎵 YouTube Audio Player (ytac)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  ytac <url>          Play audio from YouTube URL")
	fmt.Println("  ytac <search terms> Search and play")
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
	fmt.Println("  q			 Stop and return to menu")
	fmt.Println("  m             Mute/Unmute")
	fmt.Println("  9/0           Volume down/up (2%)")
	fmt.Println("  [/]           Speed down/up")
	fmt.Println("  Backspace     Reset speed to 1.0x")
	fmt.Println("  r             Restart from beginning")
	fmt.Println("  l             Loop entire track (toggle)")
	fmt.Println("  a             A-B loop (set start, end, clear)")
}

func searchYouTube(query string, searchCount int) []string {
	search := run("yt-dlp",
		fmt.Sprintf("ytsearch%d:%s", searchCount, query),
		"--no-warnings",
		"--print",
		"%(title)s | %(uploader)s | %(duration_string)s | %(webpage_url)s",
	)
	search.Stderr = nil

	pipe, err := search.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := search.Start(); err != nil {
		log.Fatal(err)
	}

	results := []string{}
	scanner := bufio.NewScanner(pipe)
	count := 0

	for scanner.Scan() {
		line := scanner.Text()
		results = append(results, line)
		count++

		parts := strings.Split(line, " | ")
		if len(parts) >= 3 {
			fmt.Printf("  %d. 🎵 %s\n", count, parts[0])
			fmt.Printf("     👤 %s  ⏱️  %s\n\n", parts[1], parts[2])
		}
	}

	search.Wait()
	return results
}

func playDirectURL(url string) {
	fmt.Println("▶️  Playing from URL...")
	fmt.Println("⌨️  Controls: Space=Pause | ←→=Seek | ↑↓=Volume | q=Quit")
	cmd := run("mpv",
		"--no-video",
		"--ytdl-format=bestaudio",
		"--term-osd-bar",
		"--term-playing-msg=▶️  Playing: 🎵 ${media-title}\n",
		"--input-default-bindings=no",
		"--input-builtin-bindings=yes",
		"--input-conf=/dev/stdin",
		url,
	)
	cmd.Stdout = os.Stdout
	cmd.Stdin = strings.NewReader(getKeyBindings())
	cmd.Run()
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		os.Exit(0)
	}

	ensureDependencies()

	// Configuration
	searchCount := 5
	showVerbose := false
	query := strings.Join(os.Args[1:], " ")

	// Direct URL playback
	if strings.HasPrefix(query, "http") {
		playDirectURL(query)
		return
	}

	// YouTube search
	fmt.Printf("🔍 Searching YouTube for: \"%s\"\n\n", query)
	results := searchYouTube(query, searchCount)

	if len(results) == 0 {
		fmt.Println("❌ No results found")
		return
	}

	fmt.Printf("✅ Found %d results\n", len(results))

	// Playback loop
	for {
		if !selectAndPlay(results, showVerbose) {
			break
		}
		fmt.Println("\n🔄 Returning to menu...")
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("👋 Goodbye!")
}
