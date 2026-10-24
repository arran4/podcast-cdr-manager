package podcast_cdr_manager

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

// xorrisoImageWriter is an optional externally-mastered ISO backend. It
// creates Rock Ridge and Joliet directory trees, preserving long MP3 names.
// xorriso is invoked directly (never through a shell).
type xorrisoImageWriter struct {
	stagingDir string
}

// CollisionKey reflects portable case-insensitive directory lookup while
// retaining full Rock Ridge and Joliet names (unlike the legacy writer).
func (w *xorrisoImageWriter) CollisionKey(name string) string {
	return strings.ToLower(path.Clean(name))
}

func newXorrisoImageWriter() (isoImageWriter, error) {
	if _, err := exec.LookPath("xorriso"); err != nil {
		return nil, fmt.Errorf("xorriso ISO backend requires the xorriso executable: %w", err)
	}
	dir, err := os.MkdirTemp("", "podcast-cdr-xorriso-")
	if err != nil {
		return nil, fmt.Errorf("create xorriso staging directory: %w", err)
	}
	return &xorrisoImageWriter{stagingDir: dir}, nil
}

func (w *xorrisoImageWriter) targetPath(name string) (string, error) {
	if w.stagingDir == "" {
		return "", fmt.Errorf("xorriso staging directory has been cleaned up")
	}
	// Only allow logical relative paths. Always reject backslash path
	// separators, including when executing on Unix.
	if name == "" || strings.ContainsAny(name, "\\\x00") || path.IsAbs(name) {
		return "", fmt.Errorf("invalid ISO target path %q", name)
	}
	clean := path.Clean(name)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("ISO target escapes staging directory: %q", name)
	}
	target := filepath.Join(w.stagingDir, filepath.FromSlash(clean))
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return "", fmt.Errorf("create staging parent: %w", err)
	}
	if _, err := os.Lstat(target); err == nil {
		return "", fmt.Errorf("ISO target %q already exists: %w", name, os.ErrExist)
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("inspect staging target %q: %w", name, err)
	}
	return target, nil
}

func (w *xorrisoImageWriter) AddFile(source io.Reader, name string) error {
	target, err := w.targetPath(name)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, source)
	closeErr := f.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(target)
		if copyErr != nil {
			return fmt.Errorf("copy %q into ISO staging: %w", name, copyErr)
		}
		return fmt.Errorf("close ISO staging file %q: %w", name, closeErr)
	}
	return nil
}

func (w *xorrisoImageWriter) AddLocalFile(sourcePath, name string) error {
	target, err := w.targetPath(name)
	if err != nil {
		return err
	}
	if err := os.Link(sourcePath, target); err == nil {
		return nil
	}
	f, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("open %q: %w", sourcePath, err)
	}
	defer f.Close()
	return w.AddFile(f, name)
}

func (w *xorrisoImageWriter) WriteTo(destination io.Writer, volumeName string) error {
	if w.stagingDir == "" {
		return fmt.Errorf("xorriso staging directory has been cleaned up")
	}
	output, err := os.CreateTemp("", "podcast-cdr-mastered-*.iso")
	if err != nil {
		return fmt.Errorf("create xorriso output: %w", err)
	}
	defer os.Remove(output.Name())
	if err := output.Close(); err != nil {
		return err
	}
	cmd := exec.Command("xorriso", "-as", "mkisofs", "-R", "-J",
		"-V", volumeName, "-o", output.Name(), w.stagingDir)
	if combined, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("master ISO with xorriso: %w: %s", err, strings.TrimSpace(string(combined)))
	}
	src, err := os.Open(output.Name())
	if err != nil {
		return fmt.Errorf("open mastered ISO: %w", err)
	}
	defer src.Close()
	if _, err := io.Copy(destination, src); err != nil {
		return fmt.Errorf("write mastered ISO: %w", err)
	}
	return nil
}

func (w *xorrisoImageWriter) Cleanup() error {
	if w.stagingDir == "" {
		return nil
	}
	dir := w.stagingDir
	w.stagingDir = ""
	return os.RemoveAll(dir)
}
