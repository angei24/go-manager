package mod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/angei24/go-manager/internal/config"
)

// ReadGoModVersion returns the Go version from the go directive in go.mod.
func ReadGoModVersion(dir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "go ") {
			continue
		}
		ver := strings.TrimSpace(strings.TrimPrefix(line, "go "))
		if ver == "" {
			continue
		}
		// ignore toolchain pseudo-versions on go line if any
		if i := strings.Index(ver, " "); i >= 0 {
			ver = ver[:i]
		}
		return ver, nil
	}
	return "", fmt.Errorf("no go directive in go.mod")
}

// SetGoModVersion updates go (and toolchain if present) in go.mod to match version (e.g. 1.26.3).
func SetGoModVersion(dir, version string) error {
	version = strings.TrimSpace(version)
	version = strings.TrimPrefix(version, "go")
	if version == "" {
		return fmt.Errorf("empty go version")
	}

	path := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read go.mod: %w", err)
	}

	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	var out []string
	goSet := false
	insertGoAfter := -1

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "module ") && insertGoAfter < 0 {
			insertGoAfter = len(out)
		}
		if strings.HasPrefix(trimmed, "go ") {
			out = append(out, "go "+version)
			goSet = true
			continue
		}
		if strings.HasPrefix(trimmed, "toolchain ") {
			out = append(out, "toolchain go"+version)
			continue
		}
		out = append(out, line)
	}

	if !goSet {
		goLine := "go " + version
		if insertGoAfter >= 0 && insertGoAfter <= len(out) {
			out = append(out[:insertGoAfter+1], append([]string{goLine}, out[insertGoAfter+1:]...)...)
		} else {
			out = append([]string{goLine}, out...)
		}
	}

	result := strings.Join(out, "\n") + "\n"
	return os.WriteFile(path, []byte(result), 0o644)
}

// PinProjectGoVersion sets go.mod and .gm-version to the same version.
func PinProjectGoVersion(dir, version string) error {
	if err := SetGoModVersion(dir, version); err != nil {
		return err
	}
	return config.WriteProjectVersion(dir, version)
}
