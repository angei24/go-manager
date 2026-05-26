package gover

import (
	"fmt"
	"os"

	"github.com/angei24/go-manager/internal/config"
)

// List prints installed and available Go versions.
func List(verbose bool) error {
	installed, err := DiscoverInstalled()
	if err != nil {
		return err
	}

	installedSet := make(map[string]bool)
	for _, sdk := range installed {
		installedSet[sdk.Version] = true
	}

	globalCfg, _ := config.LoadGlobal()
	globalDisplay := ""
	if globalCfg.DefaultVersion != "" {
		globalDisplay, _ = ParseUserVersion(globalCfg.DefaultVersion)
	}

	fmt.Println("Installed:")
	if len(installed) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, sdk := range installed {
			marker := ""
			if sdk.Version == globalDisplay || globalCfg.DefaultVersion == trimGo(sdk.Version) {
				marker = " (global default)"
			}
			src := sdk.Source
			if verbose {
				fmt.Printf("  %s  [%s] %s\n", sdk.Version, src, sdk.GOROOT)
			} else {
				fmt.Printf("  %s  [%s]%s\n", sdk.Version, src, marker)
			}
		}
	}

	policy, err := FetchSupportedReleases()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not fetch available versions: %v\n", err)
		return nil
	}

	fmt.Println("\nAvailable (latest stable minor versions):")
	for _, r := range policy.LatestByMinor {
		mark := ""
		if installedSet[r.Version] {
			mark = " [installed]"
		}
		fmt.Printf("  %s%s\n", r.Version, mark)
	}
	if verbose {
		fmt.Printf("\nInstallable range: %s\n", policy.rangeDescription())
		fmt.Printf("%d patch releases in supported minors\n", len(policy.Installable))
	}
	return nil
}

func trimGo(v string) string {
	if len(v) > 2 && v[:2] == "go" {
		return v[2:]
	}
	return v
}
