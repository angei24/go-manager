package gover

import "testing"

func TestParseGoVersionOutput(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"go version go1.26.3 darwin/arm64", "go1.26.3"},
		{"go version go1.24.13 darwin/arm64", "go1.24.13"},
	}
	for _, tt := range tests {
		got, err := parseGoVersionOutput(tt.in)
		if err != nil {
			t.Fatalf("parseGoVersionOutput(%q): %v", tt.in, err)
		}
		if got != tt.want {
			t.Errorf("got %q want %q", got, tt.want)
		}
	}
}

func TestClassifySource(t *testing.T) {
	if classifySource("/opt/homebrew/bin/go", "/opt/homebrew/Cellar/go/1.26.3/libexec", false) != "homebrew" {
		t.Error("expected homebrew")
	}
	if classifySource("/usr/local/go/bin/go", "/usr/local/go", false) != "official" {
		t.Error("expected official")
	}
}
