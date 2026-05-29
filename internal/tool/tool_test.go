package tool

import "testing"

func TestParseGoVersionLine(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"/Users/angei24/go/bin/dlv: go1.26.2", "go1.26.2"},
		{"/Users/angei24/go/bin/wire: go1.26.2", "go1.26.2"},
		{"C:\\Users\\angei24\\go\\bin\\dlv.exe: go1.26.2", "go1.26.2"},
	}
	for _, tt := range tests {
		m := goVersionLineRE.FindStringSubmatch(tt.line)
		if len(m) < 2 {
			t.Fatalf("no match: %q", tt.line)
		}
		if m[1] != tt.want {
			t.Errorf("line %q: got %q want %q", tt.line, m[1], tt.want)
		}
	}
}
