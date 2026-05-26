package mod

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/angei24/go-manager/internal/runtime"
)

// Sync runs go mod tidy and go mod download.
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

	if check {
		args := []string{"mod", "tidy", "-diff"}
		if verbose {
			fmt.Printf("go %v\n", args)
		}
		if err := runtime.ExecGo(dir, ver, nil, args...); err != nil {
			return fmt.Errorf("go mod tidy -diff: dependencies out of sync (run: gm sync)")
		}
		return nil
	}

	for _, args := range [][]string{
		{"mod", "tidy"},
		{"mod", "download"},
	} {
		if verbose {
			fmt.Printf("go %v\n", args)
		}
		if err := runtime.ExecGo(dir, ver, nil, args...); err != nil {
			return fmt.Errorf("go %v: %w", args, err)
		}
	}
	fmt.Println("Dependencies synced.")
	return nil
}

// RunGoModInit is used by project init.
func RunGoModInit(dir, module, goVersion string) error {
	args := []string{"mod", "init", module}
	if err := runtime.ExecGo(dir, goVersion, nil, args...); err != nil {
		// fallback: system go
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
