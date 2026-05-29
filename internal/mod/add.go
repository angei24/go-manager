package mod

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/angei24/go-manager/internal/runtime"
)

// Add runs go get for one or more packages.
func Add(pkgs []string, upgrade, verbose bool) error {
	if len(pkgs) == 0 {
		return fmt.Errorf("no packages specified")
	}
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
	args = append(args, pkgs...)
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

// Remove drops one or more module requirements via go mod edit.
func Remove(pkgs []string, verbose bool) error {
	if len(pkgs) == 0 {
		return fmt.Errorf("no packages specified")
	}
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
	args := []string{"mod", "edit"}
	for _, pkg := range pkgs {
		args = append(args, "-droprequire="+modulePathFromPkg(pkg))
	}
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
