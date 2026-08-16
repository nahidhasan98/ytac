package player

import (
	"reflect"
	"testing"

	"github.com/nahidhasan98/ytac/internal/youtube"
)

func TestBuildMPVArgs(t *testing.T) {
	args := buildMPVArgs(1, []string{"https://example.com/one", "https://example.com/two"}, false)

	wantContains := []string{
		"--no-video",
		"--ytdl-format=bestaudio",
		"--script-opts=ytdl_hook-ytdl_path=yt-dlp",
		"--msg-level=all=no,term-msg=info,statusline=status",
		"--input-conf=/dev/stdin",
		"--playlist-start=1",
		"https://example.com/one",
		"https://example.com/two",
	}
	for _, want := range wantContains {
		if !contains(args, want) {
			t.Fatalf("args missing %q: %#v", want, args)
		}
	}
}

func TestBuildMPVArgsVerbose(t *testing.T) {
	args := buildMPVArgs(0, []string{"https://example.com/one"}, true)
	if contains(args, "--msg-level=all=no,term-msg=info,statusline=status") {
		t.Fatalf("verbose args should not suppress logs: %#v", args)
	}
}

func TestExtractURLsFromResults(t *testing.T) {
	results := []youtube.Result{
		{URL: "https://example.com/one"},
		{URL: "https://example.com/two"},
	}
	want := []string{"https://example.com/one", "https://example.com/two"}

	if got := extractURLsFromResults(results); !reflect.DeepEqual(got, want) {
		t.Fatalf("urls = %#v, want %#v", got, want)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
