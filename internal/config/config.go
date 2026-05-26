package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// GlobalConfig is stored in config.toml.
type GlobalConfig struct {
	DefaultVersion string `toml:"default_version"`
}

// LoadGlobal reads global config; missing file returns zero value.
func LoadGlobal() (GlobalConfig, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return GlobalConfig{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return GlobalConfig{}, nil
		}
		return GlobalConfig{}, err
	}
	var cfg GlobalConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return GlobalConfig{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

// SaveGlobal writes global config, creating parent dirs.
func SaveGlobal(cfg GlobalConfig) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ReadProjectVersion reads .gm-version from dir or ancestors.
func ReadProjectVersion(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	for {
		p := filepath.Join(dir, ProjectVersionFile)
		data, err := os.ReadFile(p)
		if err == nil {
			v := strings.TrimSpace(string(data))
			if v != "" {
				return v, nil
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", nil
}

// WriteProjectVersion writes .gm-version in dir.
func WriteProjectVersion(dir, version string) error {
	p := filepath.Join(dir, ProjectVersionFile)
	return os.WriteFile(p, []byte(version+"\n"), 0o644)
}
