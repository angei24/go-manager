package gover

import "testing"

func TestParseUserVersion(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"1.22.5", "go1.22.5"},
		{"go1.22.5", "go1.22.5"},
		{"1.22", "go1.22.0"},
	}
	for _, tt := range tests {
		got, err := ParseUserVersion(tt.in)
		if err != nil {
			t.Fatalf("ParseUserVersion(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("ParseUserVersion(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMatchReleasePartial(t *testing.T) {
	releases := []Release{
		{Version: "go1.22.3", Stable: true},
		{Version: "go1.22.5", Stable: true},
		{Version: "go1.23.0", Stable: true},
	}
	r, err := MatchRelease("1.22", releases)
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != "go1.22.5" {
		t.Errorf("got %s want go1.22.5", r.Version)
	}
}
