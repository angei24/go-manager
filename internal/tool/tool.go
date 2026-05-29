package tool

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/angei24/go-manager/internal/config"
	gmruntime "github.com/angei24/go-manager/internal/runtime"
)

var goVersionLineRE = regexp.MustCompile(`:\s*(go[\d.]+(?:rc\d+|beta\d+)?)\s*$`)

type toolEntry struct {
	Name    string
	GoBuild string // e.g. go1.26.2, or "?" if unknown
}

// List prints executables in the Go bin directory (~/go/bin by default).
func List() error {
	binDir, err := config.GoBinDir()
	if err != nil {
		return err
	}
	entries, err := listBinaries(binDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("No tools installed (directory does not exist yet).\n")
			fmt.Printf("Bin directory: %s\n", binDir)
			fmt.Println("Install with: gm tool install <package>")
			return nil
		}
		return err
	}
	if len(entries) == 0 {
		fmt.Printf("No tools in %s\n", binDir)
		fmt.Println("Install with: gm tool install <package>")
		return nil
	}

	tools := describeTools(binDir, entries)
	fmt.Printf("Tools in %s:\n", binDir)
	for _, t := range tools {
		if t.GoBuild != "" && t.GoBuild != "?" {
			fmt.Printf("  %-12s %s\n", t.Name, t.GoBuild)
		} else {
			fmt.Printf("  %s\n", t.Name)
		}
	}
	return nil
}

func describeTools(binDir string, names []string) []toolEntry {
	paths := make([]string, len(names))
	for i, name := range names {
		paths[i] = filepath.Join(binDir, name)
	}
	versions := lookupGoBuildVersions(paths)

	out := make([]toolEntry, len(names))
	for i, name := range names {
		out[i] = toolEntry{
			Name:    name,
			GoBuild: versions[paths[i]],
		}
	}
	return out
}

func lookupGoBuildVersions(paths []string) map[string]string {
	result := make(map[string]string, len(paths))
	goBin, err := exec.LookPath("go")
	if err != nil {
		for _, p := range paths {
			result[p] = "?"
		}
		return result
	}

	args := append([]string{"version"}, paths...)
	out, err := exec.Command(goBin, args...).CombinedOutput()
	if err != nil {
		for _, p := range paths {
			result[p] = "?"
		}
		return result
	}

	for line := range strings.SplitSeq(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := goVersionLineRE.FindStringSubmatch(line)
		if len(m) < 2 {
			continue
		}
		// line: /path/to/bin: go1.26.2
		colon := strings.LastIndex(line, ":")
		if colon < 0 {
			continue
		}
		path := strings.TrimSpace(line[:colon])
		result[path] = m[1]
	}
	for _, p := range paths {
		if _, ok := result[p]; !ok {
			result[p] = "?"
		}
	}
	return result
}

func listBinaries(binDir string) ([]string, error) {
	dirEntries, err := os.ReadDir(binDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range dirEntries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if isHiddenOrNonTool(name) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if !isToolBinary(name, info.Mode()) {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

func isHiddenOrNonTool(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch filepath.Ext(name) {
	case ".sum", ".mod", ".txt", ".md":
		return true
	}
	return false
}

func isToolBinary(name string, mode os.FileMode) bool {
	if runtime.GOOS == "windows" {
		return strings.HasSuffix(strings.ToLower(name), ".exe")
	}
	return mode&0o111 != 0
}

// Install installs a tool via go install into ~/go/bin (or GOBIN).
func Install(pkg string, verbose bool) error {
	module, version := splitPkgVersion(pkg)
	binDir, err := config.GoBinDir()
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
	goVer, err := gmruntime.ResolveVersion(dir)
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
	if err := gmruntime.ExecGo(dir, goVer, env, args...); err != nil {
		return fmt.Errorf("go install: %w", err)
	}

	name := binaryName(module)
	binPath := filepath.Join(binDir, name)
	if _, err := os.Stat(binPath); err != nil {
		name = filepath.Base(module)
		binPath = filepath.Join(binDir, name)
	}
	if goVer != "" {
		fmt.Printf("Installed %s with Go %s -> %s\n", installPkg, goVer, binPath)
		fmt.Println("Note: reinstall tools after switching Go major versions if you see runtime errors.")
	} else {
		fmt.Printf("Installed %s -> %s\n", installPkg, binPath)
	}
	return nil
}

// Uninstall removes a binary from the Go bin directory.
func Uninstall(name string) error {
	binDir, err := config.GoBinDir()
	if err != nil {
		return err
	}
	binPath := filepath.Join(binDir, name)
	if err := os.Remove(binPath); err != nil {
		if os.IsNotExist(err) {
			if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(name), ".exe") {
				alt := binPath + ".exe"
				if err2 := os.Remove(alt); err2 == nil {
					fmt.Printf("Uninstalled %s\n", alt)
					return nil
				}
			}
			return fmt.Errorf("tool %q not found in %s", name, binDir)
		}
		return fmt.Errorf("remove binary: %w", err)
	}
	fmt.Printf("Uninstalled %s\n", binPath)
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
	if i := strings.LastIndex(module, "/"); i >= 0 {
		return module[i+1:]
	}
	return module
}
