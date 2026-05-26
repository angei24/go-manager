package gover

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const defaultDLAPI = "https://go.dev/dl/?mode=json"

// Release from go.dev/dl JSON API.
type Release struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []File `json:"files"`
}

// File is a downloadable Go release artifact.
type File struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	SHA256   string `json:"sha256"`
	Kind     string `json:"kind"`
}

// FetchReleases downloads stable release list from go.dev.
func FetchReleases() ([]Release, error) {
	url := defaultDLAPI
	if base := os.Getenv("GM_GO_DOWNLOAD_API"); base != "" {
		url = base
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch releases: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("fetch releases: HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("parse releases: %w", err)
	}
	return releases, nil
}

// SelectArchive picks the archive file for current OS/arch.
func SelectArchive(files []File) (File, error) {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	// map go arch names to dl API
	archMap := map[string]string{
		"amd64": "amd64",
		"arm64": "arm64",
		"386":   "386",
		"arm":   "armv6l",
	}
	arch, ok := archMap[goarch]
	if !ok {
		arch = goarch
	}

	var candidates []File
	for _, f := range files {
		if f.OS != goos || f.Arch != arch {
			continue
		}
		if f.Kind != "archive" {
			continue
		}
		name := strings.ToLower(f.Filename)
		if strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip") {
			candidates = append(candidates, f)
		}
	}
	if len(candidates) == 0 {
		return File{}, fmt.Errorf("no archive for %s/%s", goos, arch)
	}
	// prefer .tar.gz on unix
	for _, f := range candidates {
		if strings.HasSuffix(f.Filename, ".tar.gz") {
			return f, nil
		}
	}
	return candidates[0], nil
}

// DownloadURL returns full URL for a release file.
func DownloadURL(filename string) string {
	base := os.Getenv("GM_GO_DOWNLOAD_BASE")
	if base == "" {
		base = "https://go.dev/dl/"
	}
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
	return base + filename
}
