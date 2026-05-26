package gover

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
)

// ParseUserVersion parses 1.22, 1.22.5, go1.22.5 into go1.22.5.
func ParseUserVersion(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty version")
	}
	if strings.HasPrefix(input, "go") {
		v, err := version.NewVersion(strings.TrimPrefix(input, "go"))
		if err != nil {
			return "", fmt.Errorf("invalid version %q: %w", input, err)
		}
		seg := v.Segments()
		for len(seg) < 3 {
			seg = append(seg, 0)
		}
		return fmt.Sprintf("go%d.%d.%d", seg[0], seg[1], seg[2]), nil
	}

	v, err := version.NewVersion(input)
	if err != nil {
		return "", fmt.Errorf("invalid version %q: %w", input, err)
	}
	seg := v.Segments()
	for len(seg) < 3 {
		seg = append(seg, 0)
	}
	return fmt.Sprintf("go%d.%d.%d", seg[0], seg[1], seg[2]), nil
}

// MatchRelease finds best release for partial version (e.g. 1.22 -> latest 1.22.x).
func MatchRelease(requested string, releases []Release) (Release, error) {
	canonical, err := ParseUserVersion(requested)
	if err != nil {
		return Release{}, err
	}
	reqV, err := version.NewVersion(strings.TrimPrefix(canonical, "go"))
	if err != nil {
		return Release{}, err
	}
	reqSeg := reqV.Segments()
	for len(reqSeg) < 3 {
		reqSeg = append(reqSeg, 0)
	}

	var exact *Release
	var best *Release
	var bestV *version.Version

	for i := range releases {
		r := &releases[i]
		rv, err := version.NewVersion(strings.TrimPrefix(r.Version, "go"))
		if err != nil {
			continue
		}
		if r.Version == canonical {
			exact = r
			break
		}
		rs := rv.Segments()
		for len(rs) < 3 {
			rs = append(rs, 0)
		}
		// partial: 1.22.0 requested -> latest 1.22.x
		if reqSeg[2] == 0 {
			if rs[0] == reqSeg[0] && rs[1] == reqSeg[1] {
				if best == nil || rv.GreaterThan(bestV) {
					best = r
					bestV = rv
				}
			}
		}
	}
	if exact != nil {
		return *exact, nil
	}
	if best != nil {
		return *best, nil
	}
	return Release{}, fmt.Errorf("no release found for %q", requested)
}
