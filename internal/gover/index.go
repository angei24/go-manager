package gover

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/angei24/go-manager/internal/config"
)

// VersionsIndex tracks installed SDKs.
type VersionsIndex struct {
	Installed []string `json:"installed"`
}

func loadIndex() (VersionsIndex, error) {
	path, err := config.VersionsIndexPath()
	if err != nil {
		return VersionsIndex{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return VersionsIndex{}, nil
		}
		return VersionsIndex{}, err
	}
	var idx VersionsIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return VersionsIndex{}, err
	}
	return idx, nil
}

func saveIndex(idx VersionsIndex) error {
	path, err := config.VersionsIndexPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func addToIndex(version string) error {
	idx, err := loadIndex()
	if err != nil {
		return err
	}
	for _, v := range idx.Installed {
		if v == version {
			return nil
		}
	}
	idx.Installed = append(idx.Installed, version)
	return saveIndex(idx)
}

func removeFromIndex(version string) error {
	idx, err := loadIndex()
	if err != nil {
		return err
	}
	var next []string
	for _, v := range idx.Installed {
		if v != version {
			next = append(next, v)
		}
	}
	idx.Installed = next
	return saveIndex(idx)
}

// GOROOTPath returns path to installed SDK (gm-managed or system).
func GOROOTPath(version string) (string, error) {
	canonical, err := ParseUserVersion(version)
	if err != nil {
		canonical = version
		if !hasGoPrefix(canonical) {
			canonical = "go" + canonical
		}
	}
	versionsDir, err := config.VersionsDir()
	if err == nil {
		root := filepath.Join(versionsDir, canonical)
		if _, err := os.Stat(filepath.Join(root, "bin", "go")); err == nil {
			return root, nil
		}
	}
	return FindInstalledGOROOT(canonical)
}

func hasGoPrefix(s string) bool {
	return len(s) >= 2 && s[:2] == "go"
}
