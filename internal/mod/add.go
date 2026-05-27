package mod

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/angei24/go-manager/internal/runtime"
)

// Add runs go get for a package.
func Add(pkg string, upgrade, verbose bool) error {
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
	args := []string{"get"}
	if upgrade {
		args = append(args, "-u")
	}
	args = append(args, pkg)
	if verbose {
		fmt.Printf("go %v\n", args)
	}
	if err := runtime.ExecGo(dir, ver, nil, args...); err != nil {
		return fmt.Errorf("go get: %w", err)
	}
	if ver != "" {
		if err := PinProjectGoVersion(dir, ver); err != nil {
			return fmt.Errorf("keep go.mod in sync with .gm-version: %w", err)
		}
	}
	return nil
}

// Remove drops a module requirement via go mod edit.
func Remove(pkg string, verbose bool) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := requireGoMod(dir); err != nil {
		return err
	}
	modPath := modulePathFromPkg(pkg)
	ver, err := runtime.ResolveVersion(dir)
	if err != nil {
		return err
	}
	args := []string{"mod", "edit", "-droprequire=" + modPath}
	if verbose {
		fmt.Printf("go %v\n", args)
	}
	if err := runtime.ExecGo(dir, ver, nil, args...); err != nil {
		return fmt.Errorf("go mod edit: %w", err)
	}
	return nil
}

func requireGoMod(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		return fmt.Errorf("no go.mod in %s; run: gm init", dir)
	}
	return nil
}

func modulePathFromPkg(pkg string) string {
	// strip @version
	for i := len(pkg) - 1; i >= 0; i-- {
		if pkg[i] == '@' {
			pkg = pkg[:i]
			break
		}
	}
	// use first two path elements for most modules
	return pkg
}
