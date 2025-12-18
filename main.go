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
> playlist-next
< playlist-prev
[ multiply speed 0.9091
] multiply speed 1.1
BS set speed 1.0
`
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
	fmt.Println("  q             Quit player")
	fmt.Println("  m             Mute/Unmute")
	fmt.Println("  9/0           Volume down/up (2%)")
	fmt.Println("  [/]           Speed down/up")
	fmt.Println("  Backspace     Reset speed to 1.0x")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	// Check for help flag
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		printHelp()
		os.Exit(0)
	}

	// Check and install dependencies if needed
	ensureDependencies()

	query := strings.Join(os.Args[1:], " ")

	// If URL → play directly
	if strings.HasPrefix(query, "http") {
		fmt.Println("▶️  Playing from URL...")
		fmt.Println("⌨️  Controls: Space=Pause | ←→=Seek | ↑↓=Volume | q=Quit")
		cmd := run("mpv",
			"--no-video",
			"--ytdl-format=bestaudio",
			"--input-default-bindings=no",
			"--input-builtin-bindings=yes",
			"--input-conf=/dev/stdin",
			query,
		)
		cmd.Stdout = os.Stdout
		cmd.Stdin = strings.NewReader(getKeyBindings())
		cmd.Run()
		return
	}

	fmt.Printf("🔍 Searching YouTube for: \"%s\"\n\n", query)

	// yt-dlp search - 5 results with warnings suppressed
	search := run("yt-dlp",
		"ytsearch5:"+query,
		"--no-warnings",
		"--print",
		"%(title)s | %(uploader)s | %(duration_string)s | %(webpage_url)s",
	)
	search.Stderr = nil // Suppress all stderr output

	pipe, err := search.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := search.Start(); err != nil {
		log.Fatal(err)
	}

	// Display results as they arrive
	results := []string{}
	scanner := bufio.NewScanner(pipe)
	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		results = append(results, line)
		count++

		// Parse and display each result
		parts := strings.Split(line, " | ")
		if len(parts) >= 3 {
			title := parts[0]
			uploader := parts[1]
			duration := parts[2]
			fmt.Printf("  %d. 🎵 %s\n", count, title)
			fmt.Printf("     👤 %s  ⏱️  %s\n\n", uploader, duration)
		}
	}

	search.Wait()

	if len(results) == 0 {
		fmt.Println("❌ No results found")
		return
	}

	fmt.Printf("✅ Found %d results. Opening selection menu...\n", len(results))
	time.Sleep(500 * time.Millisecond) // Brief pause before fzf

	// Enhanced fzf with better keyboard controls
	fzf := run("fzf",
		"--prompt=🎵 Select a video ▶ ",
		"--header=↑↓:navigate | Enter:play | Ctrl-c:cancel",
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
	fzf.Stdin = strings.NewReader(strings.Join(results, "\n"))
	fzf.Stderr = os.Stderr

	fzfOut, err2 := fzf.StdoutPipe()
	if err2 != nil {
		log.Fatal(err2)
	}

	if err := fzf.Start(); err != nil {
		log.Fatal(err)
	}

	fzfScanner := bufio.NewScanner(fzfOut)
	if !fzfScanner.Scan() {
		fmt.Println("\n👋 Cancelled")
		return
	}

	line := fzfScanner.Text()
	parts := strings.Split(line, " | ")
	if len(parts) < 4 {
		log.Fatal("❌ Invalid selection format")
	}

	title := parts[0]
	url := parts[len(parts)-1]

	// Wait before playing
	fzf.Wait()

	// Play selected audio
	fmt.Printf("\n▶️  Now Playing: %s\n", title)
	fmt.Println("⌨️  Controls: Space=Pause | ←→=Seek | ↑↓=Volume | q=Quit | m=Mute")
	fmt.Println()

	play := run("mpv",
		"--no-video",
		"--ytdl-format=bestaudio",
		"--term-osd-bar",
		"--term-playing-msg=🎵 ${media-title}",
		"--input-default-bindings=no",
		"--input-builtin-bindings=yes",
		"--input-conf=/dev/stdin",
		url,
	)

	play.Stdout = os.Stdout
	play.Stdin = strings.NewReader(getKeyBindings())
	if err := play.Run(); err != nil {
		log.Printf("⚠️  Playback error: %v", err)
	}

	fmt.Println("\n✅ Playback finished")
}
