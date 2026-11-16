package app

import (
	"strings"
	"testing"

	"github.com/nahidhasan98/ytac"
)

const sampleChangelog = `# Changelog

Intro text.

## [1.2.0] - 2026-12-01

### Added
- New thing.

## [1.1.0] - 2026-11-20

### Fixed
- Old thing.
`

func TestLatestVersion(t *testing.T) {
	if got := latestVersion(sampleChangelog); got != "1.2.0" {
		t.Fatalf("latestVersion = %q, want 1.2.0", got)
	}
	if got := latestVersion("# Changelog\n"); got != "dev" {
		t.Fatalf("latestVersion without entries = %q, want dev", got)
	}
}

func TestLatestEntry(t *testing.T) {
	want := "### Added\n- New thing."
	if got := latestEntry(sampleChangelog); got != want {
		t.Fatalf("latestEntry = %q, want %q", got, want)
	}
	if got := latestEntry("# Changelog\n"); got != "" {
		t.Fatalf("latestEntry without entries = %q, want empty", got)
	}
}

// The embedded CHANGELOG.md must always have a parsable newest entry.
func TestEmbeddedChangelog(t *testing.T) {
	if v := latestVersion(ytac.Changelog); v == "dev" || strings.TrimSpace(v) == "" {
		t.Fatalf("CHANGELOG.md has no version heading")
	}
	if latestEntry(ytac.Changelog) == "" {
		t.Fatalf("CHANGELOG.md newest entry is empty")
	}
}

func TestFormatEntry(t *testing.T) {
	in := "### Added\n- Use `-V` now."
	want := "Added:\n  • Use -V now."
	if got := formatEntry(in); got != want {
		t.Fatalf("formatEntry = %q, want %q", got, want)
	}
}
