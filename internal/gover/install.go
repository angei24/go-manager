package gover

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/angei24/go-manager/internal/config"
)

// Install downloads and installs a Go version.
func Install(requested string, verbose bool) error {
	policy, err := FetchSupportedReleases()
	if err != nil {
		return err
	}
	release, err := MatchRelease(requested, policy.Installable)
	if err != nil {
		return fmt.Errorf("%w (supported: %s)", err, policy.rangeDescription())
	}
	if !release.Stable {
		return fmt.Errorf("version %s is not a stable release", release.Version)
	}
	if !policy.allows(release) {
		return fmt.Errorf("version %s is outside supported range (%s)", release.Version, policy.rangeDescription())
	}
	file, err := SelectArchive(release.Files)
	if err != nil {
		return err
	}

	versionsDir, err := config.VersionsDir()
	if err != nil {
		return err
	}
	dest := filepath.Join(versionsDir, release.Version)
	if _, err := os.Stat(filepath.Join(dest, "bin", "go")); err == nil {
		fmt.Printf("Go %s already installed at %s\n", release.Version, dest)
		return nil
	}

	if verbose {
		fmt.Printf("Downloading %s...\n", file.Filename)
	}
	tmp, err := downloadFile(DownloadURL(file.Filename), file.SHA256, verbose)
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp) }()

	if err := os.MkdirAll(versionsDir, 0o755); err != nil {
		return err
	}
	stage := dest + ".staging"
	_ = os.RemoveAll(stage)
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()

	if strings.HasSuffix(file.Filename, ".zip") {
		if err := extractZip(tmp, stage); err != nil {
			return err
		}
	} else {
		if err := extractTarGz(tmp, stage); err != nil {
			return err
		}
	}

	// archives contain top-level "go/" directory
	extracted := filepath.Join(stage, "go")
	if _, err := os.Stat(extracted); err != nil {
		// maybe files are directly in stage
		extracted = stage
	}
	_ = os.RemoveAll(dest)
	if err := os.Rename(extracted, dest); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	if err := addToIndex(release.Version); err != nil {
		return err
	}
	fmt.Printf("Installed Go %s to %s\n", release.Version, dest)
	return nil
}

func downloadFile(url, wantSHA string, verbose bool) (string, error) {
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return "", fmt.Errorf("download: HTTP %d", resp.StatusCode)
	}

	tmp, err := os.CreateTemp("", "gm-go-*")
	if err != nil {
		_ = resp.Body.Close()
		return "", err
	}
	tmpPath := tmp.Name()
	h := sha256.New()
	w := io.MultiWriter(tmp, h)
	n, err := io.Copy(w, resp.Body)
	closeErr := resp.Body.Close()
	if err := tmp.Close(); err != nil && closeErr == nil {
		closeErr = err
	}
	if err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		return "", closeErr
	}
	if verbose {
		fmt.Printf("Downloaded %d bytes\n", n)
	}
	sum := hex.EncodeToString(h.Sum(nil))
	if wantSHA != "" && !strings.EqualFold(sum, wantSHA) {
		_ = os.Remove(tmpPath)
		return "", fmt.Errorf("checksum mismatch: got %s want %s", sum, wantSHA)
	}
	return tmpPath, nil
}

func extractTarGz(archive, dest string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	return extractTar(tr, dest)
}

func extractTar(tr *tar.Reader, dest string) error {
	dest = filepath.Clean(dest)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		target := filepath.Join(dest, hdr.Name)
		if !strings.HasPrefix(target, dest+string(os.PathSeparator)) && target != dest {
			return fmt.Errorf("invalid tar path: %s", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				_ = out.Close()
				return err
			}
			_ = out.Close()
		}
	}
}

func extractZip(archive, dest string) error {
	r, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	dest = filepath.Clean(dest)
	for _, f := range r.File {
		target := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(target, dest+string(os.PathSeparator)) && target != dest {
			return fmt.Errorf("invalid zip path: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, f.Mode())
		if err != nil {
			_ = rc.Close()
			return err
		}
		_, err = io.Copy(out, rc)
		_ = rc.Close()
		_ = out.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Uninstall removes a gm-managed installed version.
func Uninstall(requested string) error {
	canonical, err := ParseUserVersion(requested)
	if err != nil {
		return err
	}
	versionsDir, err := config.VersionsDir()
	if err != nil {
		return err
	}
	dest := filepath.Join(versionsDir, canonical)
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	_ = removeFromIndex(canonical)
	fmt.Printf("Uninstalled %s\n", canonical)
	return nil
}
