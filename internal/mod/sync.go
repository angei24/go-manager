package mod

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/angei24/go-manager/internal/config"
	"github.com/angei24/go-manager/internal/runtime"
)

// Sync runs go mod tidy and go mod download using the pinned .gm-version.
func Sync(check, verbose bool) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := requireGoMod(dir); err != nil {
		return err
	}
	ver, err := runtime.ResolveVersion(dir)
	if err != nil {
		return err
	}
	if ver == "" {
		return fmt.Errorf("no Go version pinned; run: gm go use <version>")
	}

	if check {
		args := []string{"mod", "tidy", "-diff", "-go=" + ver}
		if verbose {
			fmt.Printf("go %v\n", args)
		}
		if err := runtime.ExecGo(dir, ver, nil, args...); err != nil {
			return fmt.Errorf("go mod tidy -diff: dependencies out of sync (run: gm sync)")
		}
		return nil
	}

	tidyArgs := []string{"mod", "tidy", "-go=" + ver}
	if verbose {
		fmt.Printf("go %v\n", tidyArgs)
	}
	if err := runtime.ExecGo(dir, ver, nil, tidyArgs...); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	downloadArgs := []string{"mod", "download"}
	if verbose {
		fmt.Printf("go %v\n", downloadArgs)
	}
	if err := runtime.ExecGo(dir, ver, nil, downloadArgs...); err != nil {
		return fmt.Errorf("go mod download: %w", err)
	}

	if err := alignGmVersionFile(dir); err != nil {
		return err
	}

	fmt.Println("Dependencies synced.")
	return nil
}

func alignGmVersionFile(dir string) error {
	ver, err := ReadGoModVersion(dir)
	if err != nil {
		return err
	}
	return config.WriteProjectVersion(dir, ver)
}

// RunGoModInit is used by project init.
func RunGoModInit(dir, module, goVersion string) error {
	args := []string{"mod", "init", module}
	if err := runtime.ExecGo(dir, goVersion, nil, args...); err != nil {
		cmd := exec.Command("go", args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err2 := cmd.Run(); err2 != nil {
			return fmt.Errorf("go mod init: %w", err)
		}
	}
	return nil
}

// EnsureGoMod checks go.mod exists.
func EnsureGoMod(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}
