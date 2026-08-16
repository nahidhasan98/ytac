package deps

import (
	"encoding/json"
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

var promptedYTDLPNightlyUpdate bool

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

func PromptUpdateYTDLPToNightly() bool {
	if promptedYTDLPNightlyUpdate {
		fmt.Println("💡 Already tried the yt-dlp nightly recovery during this run.")
		return false
	}
	promptedYTDLPNightlyUpdate = true

	installedPath, err := exec.LookPath("yt-dlp")
	if err != nil {
		fmt.Printf("⚠️  Could not find yt-dlp for a nightly update: %v\n", err)
		return false
	}

	fmt.Println("💡 YouTube playback failed. A newer yt-dlp nightly often fixes sudden YouTube 403 errors before the next stable release.")
	if !promptYesNo("📥 Would you like to update yt-dlp to nightly and retry playback now? (y/N): ") {
		return false
	}

	if err := updateYTDLPToNightly(installedPath); err != nil {
		fmt.Printf("❌ Failed to update yt-dlp to nightly: %v\n", err)
		return false
	}

	return true
}

func updateYTDLPToNightly(installedPath string) error {
	cmd := exec.Command(installedPath, "--update-to", "nightly")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err == nil {
		return nil
	} else if runtime.GOOS == "windows" || os.Geteuid() == 0 || !checkDependency("sudo") {
		return err
	}

	fmt.Println("🔐 yt-dlp is installed in a protected location. Retrying the update with sudo...")
	sudoCmd := exec.Command("sudo", installedPath, "--update-to", "nightly")
	sudoCmd.Stdin = os.Stdin
	sudoCmd.Stdout = os.Stdout
	sudoCmd.Stderr = os.Stderr
	return sudoCmd.Run()
}
