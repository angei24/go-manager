package gover

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/angei24/go-manager/internal/config"
)

// InstalledSDK is a detected Go installation (gm-managed or system).
type InstalledSDK struct {
	Version string // canonical, e.g. go1.26.3
	GOROOT  string
	GoBin   string
	Source  string // gm, homebrew, official, system
	Managed bool
}

var goVersionLine = regexp.MustCompile(`go version go(\d+\.\d+(?:\.\d+)?(?:rc\d+|beta\d+)?)`)

// DiscoverInstalled returns all Go SDKs found on the system plus gm-managed installs.
func DiscoverInstalled() ([]InstalledSDK, error) {
	byGOROOT := make(map[string]InstalledSDK)

	for _, bin := range collectGoBinCandidates() {
		sdk, err := inspectGoBinary(bin)
		if err != nil {
			continue
		}
		if existing, ok := byGOROOT[sdk.GOROOT]; !ok || sourcePriority(sdk) > sourcePriority(existing) {
			byGOROOT[sdk.GOROOT] = sdk
		}
	}

	gmVersions, err := listGmVersionNames()
	if err != nil {
		return nil, err
	}
	versionsDir, _ := config.VersionsDir()
	for _, name := range gmVersions {
		root := filepath.Join(versionsDir, name)
		bin := filepath.Join(root, "bin", "go")
		if _, err := os.Stat(bin); err != nil {
			continue
		}
		sdk, err := inspectGoBinary(bin)
		if err != nil {
			sdk = InstalledSDK{
				Version: name,
				GOROOT:  root,
				GoBin:   bin,
				Source:  "gm",
				Managed: true,
			}
		}
		sdk.Managed = true
		sdk.Source = "gm"
		byGOROOT[sdk.GOROOT] = sdk
	}

	// One entry per version; prefer gm > official > homebrew > system.
	byVersion := make(map[string]InstalledSDK)
	for _, sdk := range byGOROOT {
		if cur, ok := byVersion[sdk.Version]; !ok || sourcePriority(sdk) > sourcePriority(cur) {
			byVersion[sdk.Version] = sdk
		}
	}

	out := make([]InstalledSDK, 0, len(byVersion))
	for _, sdk := range byVersion {
		out = append(out, sdk)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Version > out[j].Version
	})
	return out, nil
}

func sourcePriority(s InstalledSDK) int {
	switch s.Source {
	case "gm":
		return 4
	case "official":
		return 3
	case "homebrew":
		return 2
	default:
		return 1
	}
}

func listGmVersionNames() ([]string, error) {
	idx, err := loadIndex()
	if err != nil {
		return nil, err
	}
	versionsDir, err := config.VersionsDir()
	if err != nil {
		return idx.Installed, nil
	}
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return idx.Installed, nil
	}
	seen := make(map[string]bool)
	for _, v := range idx.Installed {
		seen[v] = true
	}
	names := append([]string{}, idx.Installed...)
	for _, e := range entries {
		if e.IsDir() && !seen[e.Name()] {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

func collectGoBinCandidates() []string {
	seen := make(map[string]bool)
	var bins []string

	add := func(p string) {
		p = filepath.Clean(p)
		if p == "" || seen[p] {
			return
		}
		if _, err := os.Stat(p); err != nil {
			return
		}
		seen[p] = true
		bins = append(bins, p)
	}

	if p, err := exec.LookPath("go"); err == nil {
		add(p)
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		add(filepath.Join(dir, "go"))
	}
	if goroot := os.Getenv("GOROOT"); goroot != "" {
		add(filepath.Join(goroot, "bin", "go"))
	}

	for _, p := range standardGoBinPaths() {
		add(p)
	}
	for _, p := range globGoBins() {
		add(p)
	}
	return bins
}

func standardGoBinPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		paths := []string{
			"/usr/local/go/bin/go",
			"/opt/homebrew/bin/go",
			"/opt/homebrew/opt/go/libexec/bin/go",
			"/usr/local/bin/go",
			"/usr/local/opt/go/libexec/bin/go",
		}
		if home != "" {
			paths = append(paths, filepath.Join(home, "go", "bin", "go"))
		}
		return paths
	case "linux":
		return []string{
			"/usr/local/go/bin/go",
			"/usr/lib/go/bin/go",
		}
	case "windows":
		return []string{
			filepath.Join(os.Getenv("ProgramFiles"), "Go", "bin", "go.exe"),
		}
	default:
		return []string{"/usr/local/go/bin/go"}
	}
}

func globGoBins() []string {
	patterns := []string{
		"/opt/homebrew/Cellar/go/*/libexec/bin/go",
		"/opt/homebrew/Cellar/go/*/bin/go",
		"/opt/homebrew/Cellar/go@*/*/bin/go",
		"/opt/homebrew/Cellar/go@*/*/libexec/bin/go",
		"/opt/homebrew/opt/go@*/bin/go",
		"/opt/homebrew/opt/go@*/libexec/bin/go",
		"/usr/local/Cellar/go/*/libexec/bin/go",
		"/usr/local/Cellar/go/*/bin/go",
		"/usr/local/Cellar/go@*/*/bin/go",
		"/usr/local/opt/go@*/bin/go",
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		patterns = append(patterns, filepath.Join(home, "sdk", "go*", "bin", "go"))
	}
	var out []string
	for _, pat := range patterns {
		matches, _ := filepath.Glob(pat)
		out = append(out, matches...)
	}
	return out
}

func inspectGoBinary(goBin string) (InstalledSDK, error) {
	goBin = filepath.Clean(goBin)
	goroot := inferGOROOTFromBin(goBin)
	version, verr := readGOROOTVersion(goroot)
	if verr != nil {
		// Fallback when layout is non-standard (e.g. some wrappers).
		gorootBytes, err := exec.Command(goBin, "env", "GOROOT").Output()
		if err != nil {
			return InstalledSDK{}, err
		}
		goroot = strings.TrimSpace(string(gorootBytes))
		verBytes, err := exec.Command(goBin, "version").Output()
		if err != nil {
			return InstalledSDK{}, err
		}
		version, err = parseGoVersionOutput(string(verBytes))
		if err != nil {
			return InstalledSDK{}, err
		}
	}
	if !strings.HasPrefix(version, "go") {
		version = "go" + version
	}
	canonical, err := ParseUserVersion(strings.TrimPrefix(version, "go"))
	if err == nil {
		version = canonical
	}

	versionsDir, _ := config.VersionsDir()
	managed := versionsDir != "" && strings.HasPrefix(filepath.Clean(goroot), filepath.Clean(versionsDir))

	return InstalledSDK{
		Version: version,
		GOROOT:  goroot,
		GoBin:   goBin,
		Source:  classifySource(goBin, goroot, managed),
		Managed: managed,
	}, nil
}

func inferGOROOTFromBin(goBin string) string {
	dir := filepath.Dir(goBin)
	if filepath.Base(dir) != "bin" {
		return ""
	}
	return filepath.Dir(dir)
}

func readGOROOTVersion(goroot string) (string, error) {
	if goroot == "" {
		return "", os.ErrNotExist
	}
	data, err := os.ReadFile(filepath.Join(goroot, "VERSION"))
	if err != nil {
		return "", err
	}
	line := strings.TrimSpace(strings.Split(string(data), "\n")[0])
	if line == "" {
		return "", fmt.Errorf("empty VERSION in %s", goroot)
	}
	if !strings.HasPrefix(line, "go") {
		line = "go" + line
	}
	return line, nil
}

func parseGoVersionOutput(out string) (string, error) {
	m := goVersionLine.FindStringSubmatch(out)
	if len(m) < 2 {
		return "", fmt.Errorf("parse go version from %q", strings.TrimSpace(out))
	}
	v := m[1]
	// normalize to three components for canonical go1.x.y
	parsed, err := ParseUserVersion(v)
	if err != nil {
		return "go" + v, nil
	}
	return parsed, nil
}

func classifySource(goBin, goroot string, managed bool) string {
	if managed {
		return "gm"
	}
	lower := strings.ToLower(goBin + goroot)
	switch {
	case strings.Contains(lower, "homebrew") || strings.Contains(lower, "/cellar/"):
		return "homebrew"
	case goroot == "/usr/local/go" || strings.HasPrefix(goroot, "/usr/local/go/"):
		return "official"
	default:
		return "system"
	}
}

// FindInstalledGOROOT returns GOROOT for a version from any discovered install.
func FindInstalledGOROOT(version string) (string, error) {
	canonical, err := ParseUserVersion(version)
	if err != nil {
		canonical = version
		if !hasGoPrefix(canonical) {
			canonical = "go" + canonical
		}
	}
	sdkList, err := DiscoverInstalled()
	if err != nil {
		return "", err
	}
	for _, sdk := range sdkList {
		if sdk.Version == canonical {
			return sdk.GOROOT, nil
		}
	}
	// partial match: 1.22 -> newest installed 1.22.x
	partial, err := MatchRelease(strings.TrimPrefix(canonical, "go"), releasesFromSDKs(sdkList))
	if err != nil {
		return "", os.ErrNotExist
	}
	for _, sdk := range sdkList {
		if sdk.Version == partial.Version {
			return sdk.GOROOT, nil
		}
	}
	return "", os.ErrNotExist
}

func releasesFromSDKs(sdks []InstalledSDK) []Release {
	out := make([]Release, len(sdks))
	for i, sdk := range sdks {
		out[i] = Release{Version: sdk.Version, Stable: true}
	}
	return out
}
