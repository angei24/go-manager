package gover

import (
	"fmt"
	"os"
	"strings"

	"github.com/angei24/go-manager/internal/config"
)

// Use sets project or global default version. Returns the pinned display version (e.g. 1.26.3).
func Use(requested string, global bool) (string, error) {
	canonical, err := ParseUserVersion(requested)
	if err != nil {
		return "", fmt.Errorf("invalid version %q: %w", requested, err)
	}

	if _, err := GOROOTPath(canonical); err != nil {
		return "", fmt.Errorf("version %s not installed; run: gm go install %s", requested, requested)
	}

	display := strings.TrimPrefix(canonical, "go")
	if global {
		if err := config.SaveGlobal(config.GlobalConfig{DefaultVersion: display}); err != nil {
			return "", err
		}
		fmt.Printf("Global default Go version set to %s\n", display)
		return display, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if err := config.WriteProjectVersion(wd, display); err != nil {
		return "", err
	}
	return display, nil
}
