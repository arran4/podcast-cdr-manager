package podcast_cdr_manager

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestISOBackendSelectionDefaultsToDiskFS(t *testing.T) {
	t.Setenv("PODCAST_CDR_ISO_BACKEND", "")
	writer, err := newISOImageWriter()
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Cleanup(); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PODCAST_CDR_ISO_BACKEND", "diskfs")
	writer, err = newISOImageWriter()
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Cleanup(); err != nil {
		t.Fatal(err)
	}
}

func TestISOBackendSelectionRejectsUnknownBackend(t *testing.T) {
	t.Setenv("PODCAST_CDR_ISO_BACKEND", "not-a-backend")
	if _, err := newISOImageWriter(); err == nil || !strings.Contains(err.Error(), "unknown ISO backend") {
		t.Fatalf("expected unknown ISO backend error, got %v", err)
	}
}

func TestISOBackendSelectionXorrisoRequiresExecutable(t *testing.T) {
	t.Setenv("PODCAST_CDR_ISO_BACKEND", "xorriso")
	if _, err := exec.LookPath("xorriso"); err != nil {
		if _, writerErr := newISOImageWriter(); writerErr == nil {
			t.Fatal("expected missing xorriso executable error")
		}
		return
	}
	writer, err := newISOImageWriter()
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Cleanup(); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}
