package gover

import "testing"

func TestApplySupportedPolicy(t *testing.T) {
	releases := []Release{
		{Version: "go1.26.3", Stable: true},
		{Version: "go1.26.2", Stable: true},
		{Version: "go1.25.10", Stable: true},
		{Version: "go1.25.0", Stable: true},
		{Version: "go1.24.13", Stable: true},
		{Version: "go1.27rc1", Stable: false},
	}
	p, err := ApplySupportedPolicy(releases)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Minors) != 2 {
		t.Fatalf("minors: got %d want 2", len(p.Minors))
	}
	if p.Minors[0].String() != "1.26" || p.Minors[1].String() != "1.25" {
		t.Fatalf("minors: %+v", p.Minors)
	}
	if len(p.LatestByMinor) != 2 {
		t.Fatalf("latest: %d", len(p.LatestByMinor))
	}
	if p.LatestByMinor[0].Version != "go1.26.3" {
		t.Errorf("latest 1.26: %s", p.LatestByMinor[0].Version)
	}
	for _, r := range p.Installable {
		if r.Version == "go1.24.13" {
			t.Error("1.24 should not be installable")
		}
	}
}
