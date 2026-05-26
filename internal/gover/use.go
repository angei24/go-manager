package gover

import (
	"fmt"
	"os"
	"strings"

	"github.com/angei24/go-manager/internal/config"
)

// Use sets project or global default version.
func Use(requested string, global bool) error {
	canonical := requested
	if policy, err := FetchSupportedReleases(); err == nil {
		if r, err := MatchRelease(requested, policy.Installable); err == nil {
			canonical = r.Version
		} else if c, perr := ParseUserVersion(requested); perr == nil {
			canonical = c
		}
	} else if c, perr := ParseUserVersion(requested); perr == nil {
		canonical = c
	}

	if _, err := GOROOTPath(canonical); err != nil {
		return fmt.Errorf("version %s not installed; run: gm go install %s", requested, requested)
	}

	display := strings.TrimPrefix(canonical, "go")
	if global {
		if err := config.SaveGlobal(config.GlobalConfig{DefaultVersion: display}); err != nil {
			return err
		}
		fmt.Printf("Global default Go version set to %s\n", display)
		return nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := config.WriteProjectVersion(wd, display); err != nil {
		return err
	}
	fmt.Printf("Project Go version set to %s (.gm-version)\n", display)
	return nil
}
