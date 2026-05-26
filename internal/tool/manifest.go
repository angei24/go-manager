package tool

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/angei24/go-manager/internal/config"
)

// Entry records an installed tool.
type Entry struct {
	Name        string    `json:"name"`
	Module      string    `json:"module"`
	Version     string    `json:"version,omitempty"`
	InstalledAt time.Time `json:"installed_at"`
}

type manifest struct {
	Tools []Entry `json:"tools"`
}

var manifestMu sync.Mutex

func loadManifest() (manifest, error) {
	path, err := config.ToolsManifestPath()
	if err != nil {
		return manifest{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return manifest{}, nil
		}
		return manifest{}, err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return manifest{}, err
	}
	return m, nil
}

func saveManifest(m manifest) error {
	path, err := config.ToolsManifestPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	manifestMu.Lock()
	defer manifestMu.Unlock()
	return os.WriteFile(path, data, 0o644)
}

func upsertEntry(e Entry) error {
	m, err := loadManifest()
	if err != nil {
		return err
	}
	for i, t := range m.Tools {
		if t.Name == e.Name {
			m.Tools[i] = e
			return saveManifest(m)
		}
	}
	m.Tools = append(m.Tools, e)
	return saveManifest(m)
}

func removeEntry(name string) error {
	m, err := loadManifest()
	if err != nil {
		return err
	}
	var next []Entry
	found := false
	for _, t := range m.Tools {
		if t.Name == name {
			found = true
			continue
		}
		next = append(next, t)
	}
	if !found {
		return os.ErrNotExist
	}
	m.Tools = next
	return saveManifest(m)
}
