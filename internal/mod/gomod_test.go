package mod

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadGoModVersion(t *testing.T) {
	dir := t.TempDir()
	content := "module example.com/app\n\ngo 1.24.13\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadGoModVersion(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.24.13" {
		t.Errorf("got %q want 1.24.13", got)
	}
}

func TestSetGoModVersion(t *testing.T) {
	dir := t.TempDir()
	content := "module example.com/app\n\ngo 1.24.13\n\ntoolchain go1.24.13\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SetGoModVersion(dir, "1.26.3"); err != nil {
		t.Fatal(err)
	}
	got, err := ReadGoModVersion(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got != "1.26.3" {
		t.Errorf("go directive: got %q want 1.26.3", got)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
	if !strings.Contains(string(data), "toolchain go1.26.3") {
		t.Errorf("go.mod missing toolchain go1.26.3:\n%s", data)
	}
}

func TestPinProjectGoVersion(t *testing.T) {
	dir := t.TempDir()
	content := "module example.com/app\n\ngo 1.24.13\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := PinProjectGoVersion(dir, "1.26.3"); err != nil {
		t.Fatal(err)
	}
	gm, err := os.ReadFile(filepath.Join(dir, ".gm-version"))
	if err != nil {
		t.Fatal(err)
	}
	if string(gm) != "1.26.3\n" {
		t.Errorf(".gm-version: %q", gm)
	}
}
