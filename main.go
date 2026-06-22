package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	ytDLPRepoAPIURL            = "https://api.github.com/repos/yt-dlp/yt-dlp/releases/latest"
	ytDLPReleaseAPITimeout     = 20 * time.Second
	ytDLPBinaryDownloadTimeout = 5 * time.Minute
)

type gitHubRelease struct {
	TagName string               `json:"tag_name"`
	Assets  []gitHubReleaseAsset `json:"assets"`
}

type gitHubReleaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

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

func promptYesNo(message string) bool {
	fmt.Print(message)

	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "y" || response == "yes"
}

func fetchLatestYTDLPRelease() (gitHubRelease, error) {
	client := &http.Client{Timeout: ytDLPReleaseAPITimeout}
	req, err := http.NewRequest(http.MethodGet, ytDLPRepoAPIURL, nil)
	if err != nil {
		return gitHubRelease{}, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ytac")

	resp, err := client.Do(req)
	if err != nil {
		return gitHubRelease{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return gitHubRelease{}, fmt.Errorf("GitHub API returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var release gitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return gitHubRelease{}, err
	}
	if release.TagName == "" {
		return gitHubRelease{}, fmt.Errorf("latest yt-dlp release did not include a tag name")
	}

	return release, nil
}

func pickYTDLPAssetName() (string, error) {
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "yt-dlp.exe", nil
		}
		return "", fmt.Errorf("unsupported Windows architecture %q for yt-dlp standalone release", runtime.GOARCH)
	case "darwin":
		return "yt-dlp_macos", nil
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "yt-dlp_linux", nil
		case "arm64":
			return "yt-dlp_linux_aarch64", nil
		default:
			return "yt-dlp", nil
		}
	case "freebsd", "openbsd", "netbsd", "dragonfly":
		return "yt-dlp", nil
	default:
		return "yt-dlp", nil
	}
}

func findYTDLPAssetURL(release gitHubRelease) (string, error) {
	assetName, err := pickYTDLPAssetName()
	if err != nil {
		return "", err
	}

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return asset.URL, nil
		}
	}

	return "", fmt.Errorf("could not find a yt-dlp asset for %s/%s (expected %q)", runtime.GOOS, runtime.GOARCH, assetName)
}

func installedYTDLPVersion(path string) (string, error) {
	output, err := exec.Command(path, "--version").Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func compareVersionPart(left, right string) int {
	leftValue, leftErr := strconv.Atoi(left)
	rightValue, rightErr := strconv.Atoi(right)
	if leftErr == nil && rightErr == nil {
		switch {
		case leftValue < rightValue:
			return -1
		case leftValue > rightValue:
			return 1
		default:
			return 0
		}
	}

	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}

func compareYTDLPVersions(currentVersion, latestVersion string) int {
	currentParts := strings.Split(currentVersion, ".")
	latestParts := strings.Split(latestVersion, ".")
	maxParts := len(currentParts)
	if len(latestParts) > maxParts {
		maxParts = len(latestParts)
	}

	for i := 0; i < maxParts; i++ {
		currentPart := "0"
		if i < len(currentParts) {
			currentPart = currentParts[i]
		}

		latestPart := "0"
		if i < len(latestParts) {
			latestPart = latestParts[i]
		}

		result := compareVersionPart(currentPart, latestPart)
		if result != 0 {
			return result
		}
	}

	return 0
}

func defaultYTDLPInstallPath() (string, error) {
	executableName := "yt-dlp"
	if runtime.GOOS == "windows" {
		executableName = "yt-dlp.exe"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(homeDir, "bin", executableName), nil
	}

	return filepath.Join("/usr/local/bin", executableName), nil
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func canWriteFile(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err == nil {
		file.Close()
		return true
	}

	if !os.IsNotExist(err) {
		return false
	}

	return canWriteDir(filepath.Dir(path))
}

func canWriteDir(path string) bool {
	if err := ensureDir(path); err != nil {
		return false
	}

	testFile, err := os.CreateTemp(path, ".ytac-write-test-*")
	if err != nil {
		return false
	}
	testFile.Close()
	os.Remove(testFile.Name())
	return true
}

func downloadFile(url, destination string) error {
	client := &http.Client{Timeout: ytDLPBinaryDownloadTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "ytac")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with %s", resp.Status)
	}

	file, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

func installDownloadedYTDLP(tempPath, targetPath string) error {
	targetDir := filepath.Dir(targetPath)
	if canWriteDir(targetDir) && canWriteFile(targetPath) {
		if err := ensureDir(targetDir); err != nil {
			return err
		}

		source, err := os.Open(tempPath)
		if err != nil {
			return err
		}
		defer source.Close()

		destination, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
		if err != nil {
			return err
		}
		defer destination.Close()

		if _, err := io.Copy(destination, source); err != nil {
			return err
		}

		return os.Chmod(targetPath, 0o755)
	}

	if runtime.GOOS == "windows" {
		return fmt.Errorf("cannot write to %s; please rerun from a shell with permission to replace it", targetPath)
	}

	cmd := exec.Command("sudo", "install", "-m", "755", tempPath, targetPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func installOrUpdateYTDLP(targetPath, downloadURL, releaseTag string, action string) error {
	tempFile, err := os.CreateTemp("", "yt-dlp-*")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	fmt.Printf("📦 %s yt-dlp %s...\n", action, releaseTag)
	if err := downloadFile(downloadURL, tempPath); err != nil {
		return err
	}

	if err := installDownloadedYTDLP(tempPath, targetPath); err != nil {
		return err
	}

	fmt.Printf("✅ yt-dlp %s at %s\n", strings.ToLower(action), targetPath)
	return nil
}

func ensureYTDLP() {
	installedPath, err := exec.LookPath("yt-dlp")
	hasInstalledYTDLP := err == nil

	release, releaseErr := fetchLatestYTDLPRelease()
	if releaseErr != nil {
		if hasInstalledYTDLP {
			fmt.Printf("⚠️  Could not check for yt-dlp updates: %v\n", releaseErr)
			return
		}
		log.Fatalf("❌ Failed to check yt-dlp releases: %v\n", releaseErr)
	}

	downloadURL, downloadErr := findYTDLPAssetURL(release)
	if downloadErr != nil {
		if hasInstalledYTDLP {
			fmt.Printf("⚠️  Could not resolve a yt-dlp download for %s/%s: %v\n", runtime.GOOS, runtime.GOARCH, downloadErr)
			return
		}
		log.Fatalf("❌ %v\n", downloadErr)
	}

	if hasInstalledYTDLP {
		currentVersion, versionErr := installedYTDLPVersion(installedPath)
		if versionErr != nil {
			fmt.Printf("⚠️  Could not read current yt-dlp version from %s: %v\n", installedPath, versionErr)
			if promptYesNo("📥 Would you like to reinstall yt-dlp from the latest release now? (y/N): ") {
				if err := installOrUpdateYTDLP(installedPath, downloadURL, release.TagName, "Reinstalling"); err != nil {
					log.Fatalf("❌ Failed to reinstall yt-dlp: %v\n", err)
				}
			}
			return
		}

		versionComparison := compareYTDLPVersions(currentVersion, release.TagName)
		if versionComparison >= 0 {
			return
		}

		fmt.Printf("⬆️  yt-dlp update available: %s -> %s\n", currentVersion, release.TagName)
		if promptYesNo("📥 Would you like to update yt-dlp now? (y/N): ") {
			if err := installOrUpdateYTDLP(installedPath, downloadURL, release.TagName, "Updating"); err != nil {
				log.Fatalf("❌ Failed to update yt-dlp: %v\n", err)
			}
		}
		return
	}

	targetPath, err := defaultYTDLPInstallPath()
	if err != nil {
		log.Fatalf("❌ Failed to determine yt-dlp install path: %v\n", err)
	}

	fmt.Printf("⚠️  Missing required dependency: yt-dlp (%s available)\n", release.TagName)
	if !promptYesNo("📥 Would you like to install yt-dlp now? (y/N): ") {
		fmt.Println("❌ Cannot proceed without yt-dlp")
		os.Exit(1)
	}

	if err := installOrUpdateYTDLP(targetPath, downloadURL, release.TagName, "Installing"); err != nil {
		log.Fatalf("❌ Failed to install yt-dlp: %v\n", err)
	}
}

func ensureDependencies() {
	ensureYTDLP()

	deps := map[string]string{
		"mpv": "mpv",
		"fzf": "fzf",
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
	if !promptYesNo("📥 Would you like to install them now? (y/N): ") {
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
	searchCount := 3
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

			value, err := strconv.Atoi(args[i])
			if err != nil {
				return "", showVerbose, 0, fmt.Errorf("invalid search count %q: must be a number between 1 and 10", args[i])
			}
			if value < 1 || value > 10 {
				return "", showVerbose, 0, fmt.Errorf("search count must be between 1 and 10")
			}
			searchCount = value
		default:
			if strings.HasPrefix(arg, "--search-count=") {
				valueText := strings.TrimPrefix(arg, "--search-count=")
				value, err := strconv.Atoi(valueText)
				if err != nil {
					return "", showVerbose, 0, fmt.Errorf("invalid search count %q: must be a number between 1 and 10", valueText)
				}
				if value < 1 || value > 10 {
					return "", showVerbose, 0, fmt.Errorf("search count must be between 1 and 10")
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

func playDirectURL(url string, showVerbose bool) {
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

	query, showVerbose, searchCount, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Printf("❌ %v\n\n", err)
		printHelp()
		os.Exit(1)
	}

	// Direct URL playback
	if strings.HasPrefix(query, "http") {
		playDirectURL(query, showVerbose)
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
