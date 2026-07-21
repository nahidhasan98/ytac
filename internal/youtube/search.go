package youtube

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
)

type Result struct {
	Title    string
	Uploader string
	Duration string
	URL      string
}

type ytDLPVideo struct {
	Title          string `json:"title"`
	Uploader       string `json:"uploader"`
	Channel        string `json:"channel"`
	DurationString string `json:"duration_string"`
	WebpageURL     string `json:"webpage_url"`
	OriginalURL    string `json:"original_url"`
}

func Search(query string, searchCount int) ([]Result, error) {
	search := exec.Command("yt-dlp",
		fmt.Sprintf("ytsearch%d:%s", searchCount, query),
		"--no-warnings",
		"--dump-json",
	)
	search.Stderr = nil

	pipe, err := search.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := search.Start(); err != nil {
		return nil, err
	}

	results := []Result{}
	decoder := json.NewDecoder(pipe)
	for {
		result, err := decodeSearchResult(decoder)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		results = append(results, result)
		fmt.Printf("  %d. 🎵 %s\n", len(results), result.Title)
		fmt.Printf("     👤 %s  ⏱️  %s\n\n", result.Uploader, result.Duration)
	}

	if err := search.Wait(); err != nil {
		return nil, err
	}

	return results, nil
}

func decodeSearchResult(decoder *json.Decoder) (Result, error) {
	var video ytDLPVideo
	if err := decoder.Decode(&video); err != nil {
		return Result{}, err
	}

	return newSearchResult(video)
}

func newSearchResult(video ytDLPVideo) (Result, error) {
	result := Result{
		Title:    fallback(video.Title, "Untitled"),
		Uploader: fallback(video.Uploader, video.Channel, "Unknown uploader"),
		Duration: fallback(video.DurationString, "Unknown duration"),
		URL:      fallback(video.WebpageURL, video.OriginalURL),
	}
	if result.URL == "" {
		return Result{}, fmt.Errorf("search result %q did not include a playable URL", result.Title)
	}

	return result, nil
}

func fallback(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
