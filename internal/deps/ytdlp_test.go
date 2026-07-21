package deps

import "testing"

func TestCompareYTDLPVersions(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    int
	}{
		{name: "equal", current: "2026.01.01", latest: "2026.01.01", want: 0},
		{name: "current older", current: "2025.12.31", latest: "2026.01.01", want: -1},
		{name: "current newer", current: "2026.02.01", latest: "2026.01.01", want: 1},
		{name: "missing patch equals zero", current: "2026.01", latest: "2026.01.0", want: 0},
		{name: "numeric compare", current: "2026.10.01", latest: "2026.2.01", want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareYTDLPVersions(tt.current, tt.latest)
			switch {
			case tt.want < 0 && got >= 0:
				t.Fatalf("compare = %d, want negative", got)
			case tt.want > 0 && got <= 0:
				t.Fatalf("compare = %d, want positive", got)
			case tt.want == 0 && got != 0:
				t.Fatalf("compare = %d, want zero", got)
			}
		})
	}
}
