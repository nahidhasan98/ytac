package ui

import (
	"reflect"
	"testing"

	"github.com/nahidhasan98/ytac/internal/youtube"
)

func TestFormatResultsForFzf(t *testing.T) {
	results := []youtube.Result{
		{Title: "Title", Uploader: "Uploader", Duration: "3:21", URL: "https://youtu.be/video"},
	}
	want := []string{"[1] Title | Uploader | 3:21"}

	if got := formatResultsForFzf(results); !reflect.DeepEqual(got, want) {
		t.Fatalf("formatted = %#v, want %#v", got, want)
	}
}
