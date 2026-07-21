package app

import "testing"

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantQuery   string
		wantVerbose bool
		wantCount   int
		wantErr     bool
	}{
		{
			name:      "search terms use default count",
			args:      []string{"lofi", "beats"},
			wantQuery: "lofi beats",
			wantCount: defaultSearchCount,
		},
		{
			name:        "verbose and short count",
			args:        []string{"-v", "-c", "10", "nasheed"},
			wantQuery:   "nasheed",
			wantVerbose: true,
			wantCount:   10,
		},
		{
			name:      "long count assignment",
			args:      []string{"--search-count=1", "tafseer"},
			wantQuery: "tafseer",
			wantCount: 1,
		},
		{
			name:    "missing query",
			args:    []string{"--verbose"},
			wantErr: true,
		},
		{
			name:    "missing count value",
			args:    []string{"--search-count"},
			wantErr: true,
		},
		{
			name:    "invalid count",
			args:    []string{"--search-count", "many", "tafseer"},
			wantErr: true,
		},
		{
			name:    "out of range count",
			args:    []string{"--search-count=11", "tafseer"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, verbose, count, err := parseArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if query != tt.wantQuery {
				t.Fatalf("query = %q, want %q", query, tt.wantQuery)
			}
			if verbose != tt.wantVerbose {
				t.Fatalf("verbose = %v, want %v", verbose, tt.wantVerbose)
			}
			if count != tt.wantCount {
				t.Fatalf("count = %d, want %d", count, tt.wantCount)
			}
		})
	}
}
