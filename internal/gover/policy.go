package gover

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
)

// SupportedMinorCount is how many latest stable minor lines gm supports (e.g. 1.25 + 1.26).
const SupportedMinorCount = 2

// SupportedPolicy describes which stable releases can be listed and installed.
type SupportedPolicy struct {
	Minors        []minorKey
	Installable   []Release
	LatestByMinor []Release
}

type minorKey struct {
	major int
	minor int
}

func (k minorKey) String() string {
	return fmt.Sprintf("%d.%d", k.major, k.minor)
}

// ApplySupportedPolicy keeps stable releases in the N latest minor versions.
func ApplySupportedPolicy(releases []Release) (SupportedPolicy, error) {
	var stable []Release
	for _, r := range releases {
		if !r.Stable {
			continue
		}
		if _, err := version.NewVersion(strings.TrimPrefix(r.Version, "go")); err != nil {
			continue
		}
		stable = append(stable, r)
	}
	if len(stable) == 0 {
		return SupportedPolicy{}, fmt.Errorf("no stable releases from API")
	}

	minors := distinctMinors(stable)
	if len(minors) == 0 {
		return SupportedPolicy{}, fmt.Errorf("no parseable stable releases")
	}
	sort.Slice(minors, func(i, j int) bool {
		return compareMinor(minors[i], minors[j]) > 0
	})
	if len(minors) > SupportedMinorCount {
		minors = minors[:SupportedMinorCount]
	}

	allowed := make(map[minorKey]bool, len(minors))
	for _, m := range minors {
		allowed[m] = true
	}

	var installable []Release
	for _, r := range stable {
		if m, ok := releaseMinor(r); ok && allowed[m] {
			installable = append(installable, r)
		}
	}
	sort.Slice(installable, func(i, j int) bool {
		return releaseGreater(installable[i].Version, installable[j].Version)
	})

	latest := make([]Release, 0, len(minors))
	for _, m := range minors {
		var best *Release
		for i := range installable {
			r := &installable[i]
			rm, ok := releaseMinor(*r)
			if !ok || rm != m {
				continue
			}
			if best == nil || releaseGreater(r.Version, best.Version) {
				best = r
			}
		}
		if best != nil {
			latest = append(latest, *best)
		}
	}

	return SupportedPolicy{
		Minors:        minors,
		Installable:   installable,
		LatestByMinor: latest,
	}, nil
}

// FetchSupportedReleases loads stable releases and applies the supported minor policy.
func FetchSupportedReleases() (SupportedPolicy, error) {
	releases, err := FetchReleases()
	if err != nil {
		return SupportedPolicy{}, err
	}
	return ApplySupportedPolicy(releases)
}

func (p SupportedPolicy) allows(release Release) bool {
	m, ok := releaseMinor(release)
	if !ok {
		return false
	}
	for _, allowed := range p.Minors {
		if m == allowed {
			return true
		}
	}
	return false
}

func (p SupportedPolicy) rangeDescription() string {
	if len(p.Minors) == 0 {
		return "no supported versions"
	}
	oldest := p.Minors[len(p.Minors)-1]
	newest := p.Minors[0]
	var maxPatch string
	for _, r := range p.Installable {
		m, ok := releaseMinor(r)
		if !ok || m != newest {
			continue
		}
		patch := strings.TrimPrefix(r.Version, "go")
		if maxPatch == "" || releaseGreater("go"+patch, "go"+maxPatch) {
			maxPatch = patch
		}
	}
	return fmt.Sprintf("%s.0 through %s (stable only)", oldest.String(), maxPatch)
}

func distinctMinors(releases []Release) []minorKey {
	seen := make(map[minorKey]bool)
	var minors []minorKey
	for _, r := range releases {
		m, ok := releaseMinor(r)
		if !ok || seen[m] {
			continue
		}
		seen[m] = true
		minors = append(minors, m)
	}
	return minors
}

func releaseMinor(r Release) (minorKey, bool) {
	v, err := version.NewVersion(strings.TrimPrefix(r.Version, "go"))
	if err != nil {
		return minorKey{}, false
	}
	seg := v.Segments()
	if len(seg) < 2 {
		return minorKey{}, false
	}
	return minorKey{seg[0], seg[1]}, true
}

func compareMinor(a, b minorKey) int {
	if a.major != b.major {
		return a.major - b.major
	}
	return a.minor - b.minor
}

func releaseGreater(a, b string) bool {
	va, err := version.NewVersion(strings.TrimPrefix(a, "go"))
	if err != nil {
		return false
	}
	vb, err := version.NewVersion(strings.TrimPrefix(b, "go"))
	if err != nil {
		return true
	}
	return va.GreaterThan(vb)
}
