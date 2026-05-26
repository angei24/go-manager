package config

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	appName = "gm"
)

// DataDir returns ~/.local/share/gm (or platform equivalent).
func DataDir() (string, error) {
	if v := os.Getenv("GM_DATA_DIR"); v != "" {
		return v, nil
	}
	base, err := userDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, appName), nil
}

// ConfigDir returns ~/.config/gm.
func ConfigDir() (string, error) {
	if v := os.Getenv("GM_CONFIG_DIR"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", appName), nil
}

// VersionsDir returns directory for installed Go SDKs.
func VersionsDir() (string, error) {
	data, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(data, "versions"), nil
}

// ToolsBinDir returns GOBIN for gm-managed tools.
func ToolsBinDir() (string, error) {
	data, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(data, "bin"), nil
}

// ToolsManifestPath returns path to tools.json.
func ToolsManifestPath() (string, error) {
	data, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(data, "tools.json"), nil
}

// VersionsIndexPath returns path to versions.json index.
func VersionsIndexPath() (string, error) {
	data, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(data, "versions.json"), nil
}

// ConfigFilePath returns path to config.toml.
func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// ProjectVersionFile is the per-project Go version pin file.
const ProjectVersionFile = ".gm-version"

func userDataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// XDG_DATA_HOME or ~/.local/share
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, appName), nil
	}
	if runtime.GOOS == "windows" {
		local := os.Getenv("LOCALAPPDATA")
		if local != "" {
			return filepath.Join(local, appName), nil
		}
	}
	return filepath.Join(home, ".local", "share", appName), nil
}
