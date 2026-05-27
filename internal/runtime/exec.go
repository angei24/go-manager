package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/angei24/go-manager/internal/config"
	"github.com/angei24/go-manager/internal/gover"
)

// ResolveVersion picks Go version: .gm-version > GM_GO_VERSION > global config > system.
func ResolveVersion(workDir string) (string, error) {
	if v := os.Getenv("GM_GO_VERSION"); v != "" {
		return v, nil
	}
	if workDir == "" {
		var err error
		workDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	if v, err := config.ReadProjectVersion(workDir); err == nil && v != "" {
		return v, nil
	}
	cfg, err := config.LoadGlobal()
	if err != nil {
		return "", err
	}
	if cfg.DefaultVersion != "" {
		return cfg.DefaultVersion, nil
	}
	return "", nil // use system go
}

// ExecGo runs the go binary for the resolved version with args in dir.
func ExecGo(workDir string, version string, env []string, args ...string) error {
	goBin, goroot, err := goBinary(version)
	if err != nil {
		return err
	}
	cmd := exec.Command(goBin, args...)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = append(os.Environ(), env...)
	if goroot != "" {
		cmd.Env = appendEnv(cmd.Env, "GOROOT="+goroot)
		cmd.Env = appendEnv(cmd.Env, "GOTOOLCHAIN=local")
		cmd.Env = prependPath(cmd.Env, filepath.Join(goroot, "bin"))
	}
	return cmd.Run()
}

// ExecGoCombined runs go and returns combined output error.
func ExecGoCombined(workDir string, version string, args ...string) ([]byte, error) {
	goBin, goroot, err := goBinary(version)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(goBin, args...)
	cmd.Dir = workDir
	env := os.Environ()
	if goroot != "" {
		env = appendEnv(env, "GOROOT="+goroot)
		env = appendEnv(env, "GOTOOLCHAIN=local")
		env = prependPath(env, filepath.Join(goroot, "bin"))
	}
	cmd.Env = env
	return cmd.CombinedOutput()
}

func goBinary(version string) (goBin, goroot string, err error) {
	if version != "" {
		canonical, perr := gover.ParseUserVersion(version)
		if perr != nil {
			canonical = version
		}
		root, err := gover.GOROOTPath(canonical)
		if err != nil {
			return "", "", fmt.Errorf("go %s not installed (gm go install %s): %w", version, version, err)
		}
		return filepath.Join(root, "bin", "go"), root, nil
	}
	path, err := exec.LookPath("go")
	if err != nil {
		return "", "", fmt.Errorf("no Go version configured and 'go' not in PATH; run: gm go install <version>")
	}
	return path, "", nil
}

func appendEnv(env []string, kv string) []string {
	key, _, ok := splitEnv(kv)
	if !ok {
		return append(env, kv)
	}
	if i := indexEnvKey(env, key); i >= 0 {
		out := make([]string, len(env))
		copy(out, env)
		out[i] = kv
		return out
	}
	return append(env, kv)
}

func indexEnvKey(env []string, kv string) int {
	key, _, _ := splitEnv(kv)
	for i, e := range env {
		k, _, ok := splitEnv(e)
		if ok && k == key {
			return i
		}
	}
	return -1
}

func splitEnv(kv string) (key, val string, ok bool) {
	for i := 0; i < len(kv); i++ {
		if kv[i] == '=' {
			return kv[:i], kv[i+1:], true
		}
	}
	return "", "", false
}

func prependPath(env []string, dir string) []string {
	const key = "PATH"
	for i, e := range env {
		if k, v, ok := splitEnv(e); ok && k == key {
			sep := ":"
			if os.PathSeparator == '\\' {
				sep = ";"
			}
			env[i] = key + "=" + dir + sep + v
			return env
		}
	}
	sep := ":"
	if os.PathSeparator == '\\' {
		sep = ";"
	}
	return append(env, key+"="+dir+sep+os.Getenv("PATH"))
}
