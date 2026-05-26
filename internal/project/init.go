package project

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/angei24/go-manager/internal/config"
	"github.com/angei24/go-manager/internal/gover"
	"github.com/angei24/go-manager/internal/mod"
)

//go:embed templates/*
var templateFS embed.FS

// InitOptions configures gm init.
type InitOptions struct {
	Dir     string
	Module  string
	NoGit   bool
	Force   bool
	Verbose bool
}

// Init scaffolds a new Go project.
func Init(opts InitOptions) error {
	dir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	entries, _ := os.ReadDir(dir)
	nonEmpty := false
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		nonEmpty = true
		break
	}
	if nonEmpty && !opts.Force {
		return fmt.Errorf("directory %s is not empty; use --force", dir)
	}

	module := opts.Module
	if module == "" {
		module = filepath.Base(dir)
		if module == "." || module == "/" {
			module = "example.com/app"
		} else {
			module = "example.com/" + module
		}
	}

	goVer := ""
	if cfg, err := config.LoadGlobal(); err == nil && cfg.DefaultVersion != "" {
		goVer = cfg.DefaultVersion
	}

	if !opts.NoGit {
		if err := runGitInit(dir, opts.Verbose); err != nil {
			return err
		}
	}

	if err := mod.RunGoModInit(dir, module, goVer); err != nil {
		return err
	}

	data := map[string]string{
		"Module": module,
	}
	if err := writeTemplate(dir, "main.go.tmpl", "main.go", data); err != nil {
		return err
	}
	if err := writeTemplate(dir, "README.md.tmpl", "README.md", data); err != nil {
		return err
	}
	if err := writeGitignore(dir); err != nil {
		return err
	}

	pinVer := goVer
	if pinVer == "" {
		if policy, err := gover.FetchSupportedReleases(); err == nil && len(policy.LatestByMinor) > 0 {
			pinVer = strings.TrimPrefix(policy.LatestByMinor[0].Version, "go")
		} else {
			pinVer = "1.22.0"
		}
	}
	if err := config.WriteProjectVersion(dir, pinVer); err != nil {
		return err
	}

	fmt.Printf("Initialized Go project in %s (module %s)\n", dir, module)
	fmt.Println("Next: gm add <package>  |  gm sync")
	return nil
}

func writeTemplate(dir, tmplFile, outName string, data any) error {
	content, err := templateFS.ReadFile(filepath.Join("templates", tmplFile))
	if err != nil {
		return err
	}
	t, err := template.New(outName).Parse(string(content))
	if err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(dir, outName))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	return t.Execute(f, data)
}

func writeGitignore(dir string) error {
	const gitignore = `# Binaries
bin/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test
*.test
*.out

# Go workspace
go.work.sum

# IDE
.idea/
.vscode/
*.swp
`
	p := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(p); err == nil {
		return nil
	}
	return os.WriteFile(p, []byte(gitignore), 0o644)
}

func runGitInit(dir string, verbose bool) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found in PATH")
	}
	gitDir := filepath.Join(dir, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		return nil
	}
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if verbose {
		fmt.Println("git init")
	}
	return cmd.Run()
}
