package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
)

func TestDiskNext_NilSizeBytes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "podcast-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	profileName := "test_profile"
	xdg.ConfigHome = tempDir

	p := &podcast_cdr_manager.Profile{
		Version: "1", Name: profileName,
		Casts: []*podcast_cdr_manager.Cast{
			{
				Title:    "Test Cast",
				MpegLink: "http://example.com/test.mp3",
				// SizeBytes is deliberately nil
			},
		},
		Disks: []*podcast_cdr_manager.Disk{},
	}
	if err := p.Save(); err != nil {
		t.Fatalf("failed to save profile: %v", err)
	}

	dedicatedIndex := -1
	create := true
	diskSizeMb := 600
	dry := false

	err = DiskNext(profileName, dedicatedIndex, diskSizeMb, create, dry)
	if err == nil {
		t.Fatalf("Expected an error when cast size bytes are missing")
	}

	if !strings.Contains(err.Error(), "size bytes missing") {
		t.Fatalf("Expected hard-fail error about missing size bytes, got: %v", err)
	}
}
