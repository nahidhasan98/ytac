package app

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nahidhasan98/ytac"
)

// Version is the ytac release version. It defaults to the newest entry in
// CHANGELOG.md, so a release only needs a new changelog entry. It can still be
// overridden at build time with:
//
//	go build -ldflags "-X github.com/nahidhasan98/ytac/internal/app.Version=x.y.z"
var Version = latestVersion(ytac.Changelog)

var versionHeading = regexp.MustCompile(`(?m)^## \[([^\]]+)\]`)

// latestVersion returns the version from the first "## [x.y.z]" heading.
func latestVersion(changelog string) string {
	if m := versionHeading.FindStringSubmatch(changelog); m != nil {
		return m[1]
	}
	return "dev"
}

// latestEntry returns the body of the newest changelog entry (without its
// heading), or "" if there is none.
func latestEntry(changelog string) string {
	locs := versionHeading.FindAllStringIndex(changelog, 2)
	if len(locs) == 0 {
		return ""
	}
	start := locs[0][1]
	if nl := strings.IndexByte(changelog[start:], '\n'); nl >= 0 {
		start += nl + 1 // skip the rest of the heading line (the date)
	} else {
		return ""
	}
	end := len(changelog)
	if len(locs) > 1 {
		end = locs[1][0]
	}
	return strings.TrimSpace(changelog[start:end])
}

func printVersion() {
	fmt.Printf("🎵 ytac v%s\n", Version)
}

// printChangelog prints what changed in the current version.
func printChangelog() {
	entry := latestEntry(ytac.Changelog)
	if entry == "" {
		return
	}
	fmt.Println()
	fmt.Printf("📝 What's new in v%s:\n\n", Version)
	fmt.Println(formatEntry(entry))
}

// formatEntry turns a markdown changelog entry into plain terminal text:
// "### Added" -> "Added:", "- item" -> "  • item", and drops backticks.
func formatEntry(entry string) string {
	lines := strings.Split(entry, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "### "):
			line = strings.TrimPrefix(line, "### ") + ":"
		case strings.HasPrefix(line, "- "):
			line = "  • " + strings.TrimPrefix(line, "- ")
		}
		lines[i] = strings.ReplaceAll(line, "`", "")
	}
	return strings.Join(lines, "\n")
}
