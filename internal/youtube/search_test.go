package youtube

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeSearchResult(t *testing.T) {
	input := `{"title":"A | title","uploader":"Uploader | Name","duration_string":"3:21","webpage_url":"https://youtu.be/video"}`

	got, err := decodeSearchResult(json.NewDecoder(strings.NewReader(input)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := Result{
		Title:    "A | title",
		Uploader: "Uploader | Name",
		Duration: "3:21",
		URL:      "https://youtu.be/video",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("result = %#v, want %#v", got, want)
	}
}

func TestNewSearchResultFallbacks(t *testing.T) {
	video := ytDLPVideo{
		Channel:     "Channel Name",
		OriginalURL: "https://youtu.be/video",
	}

	got, err := newSearchResult(video)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Title != "Untitled" {
		t.Fatalf("title = %q, want fallback", got.Title)
	}
	if got.Uploader != "Channel Name" {
		t.Fatalf("uploader = %q, want channel fallback", got.Uploader)
	}
	if got.Duration != "Unknown duration" {
		t.Fatalf("duration = %q, want fallback", got.Duration)
	}
	if got.URL != "https://youtu.be/video" {
		t.Fatalf("url = %q, want original_url fallback", got.URL)
	}
}

func TestNewSearchResultRequiresURL(t *testing.T) {
	_, err := newSearchResult(ytDLPVideo{Title: "No URL"})
	if err == nil {
		t.Fatal("expected missing URL error")
	}
}

func TestDecodeSearchResultHandlesLargeYTDLPJSON(t *testing.T) {
	video := ytDLPVideo{
		Title:          strings.Repeat("large title ", 10000),
		Uploader:       "Uploader",
		DurationString: "3:21",
		WebpageURL:     "https://youtu.be/video",
	}
	data, err := json.Marshal(video)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}

	decoder := json.NewDecoder(strings.NewReader(string(data) + "\n"))
	got, err := decodeSearchResult(decoder)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.URL != video.WebpageURL {
		t.Fatalf("url = %q, want %q", got.URL, video.WebpageURL)
	}
	if got.Title != video.Title {
		t.Fatalf("large title was not preserved")
	}
}
