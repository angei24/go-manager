package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/angei24/go-manager/internal/config"
	"github.com/angei24/go-manager/internal/runtime"
)

// List prints installed tools.
func List() error {
	m, err := loadManifest()
	if err != nil {
		return err
	}
	if len(m.Tools) == 0 {
		fmt.Println("No tools installed.")
		bin, _ := config.ToolsBinDir()
		fmt.Printf("Install with: gm tool install <package>\nBin directory: %s\n", bin)
		return nil
	}
	fmt.Println("Installed tools:")
	for _, t := range m.Tools {
		ver := t.Version
		if ver == "" {
			ver = "latest"
		}
		fmt.Printf("  %s  %s@%s\n", t.Name, t.Module, ver)
	}
	bin, _ := config.ToolsBinDir()
	fmt.Printf("\nAdd to PATH: export PATH=%s:$PATH\n", bin)
	return nil
}

// Install installs a tool via go install into gm bin dir.
func Install(pkg string, verbose bool) error {
	module, version := splitPkgVersion(pkg)
	binDir, err := config.ToolsBinDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	ver, err := runtime.ResolveVersion(dir)
	if err != nil {
		return err
	}

	installPkg := module + "@latest"
	if version != "" {
		installPkg = module + "@" + version
	}

	env := []string{"GOBIN=" + binDir}
	args := []string{"install", installPkg}
	if verbose {
		fmt.Printf("GOBIN=%s go %v\n", binDir, args)
	}
	if err := runtime.ExecGo(dir, ver, env, args...); err != nil {
		return fmt.Errorf("go install: %w", err)
	}

	name := binaryName(module)
	binPath := filepath.Join(binDir, name)
	if _, err := os.Stat(binPath); err != nil {
		// try last path segment of module
		name = filepath.Base(module)
	}

	if err := upsertEntry(Entry{
		Name:        name,
		Module:      module,
		Version:     version,
		InstalledAt: time.Now().UTC(),
	}); err != nil {
		return err
	}

	fmt.Printf("Installed %s -> %s\n", installPkg, binPath)
	fmt.Printf("Ensure PATH includes: %s\n", binDir)
	return nil
}

// Uninstall removes a tool binary and manifest entry.
func Uninstall(name string) error {
	binDir, err := config.ToolsBinDir()
	if err != nil {
		return err
	}
	binPath := filepath.Join(binDir, name)
	if err := os.Remove(binPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove binary: %w", err)
	}
	if err := removeEntry(name); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("tool %q not found in manifest", name)
		}
		return err
	}
	fmt.Printf("Uninstalled %s\n", name)
	return nil
}

func splitPkgVersion(pkg string) (module, version string) {
	at := strings.LastIndex(pkg, "@")
	if at < 0 {
		return pkg, ""
	}
	return pkg[:at], pkg[at+1:]
}

func binaryName(module string) string {
	// golang.org/x/tools/gopls -> gopls
	if i := strings.LastIndex(module, "/"); i >= 0 {
		return module[i+1:]
	}
	return module
}
