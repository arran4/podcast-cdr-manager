package main

import (
	"flag"
	"os"
	"testing"
	"strings"

	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"github.com/adrg/xdg"
)

func TestDoRunDiskNext_NilSizeBytes(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "podcast-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

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

	mc := &MainConfig{
		profile: profileName,
	}

	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	help := false
	dedicatedIndex := -1
	create := true
	diskSizeMb := 600
	dry := false

	err = DoRunDiskNext(&help, fs, mc, &dedicatedIndex, &create, &diskSizeMb, &dry)
	if err == nil {
		t.Fatalf("Expected an error since example.com/test.mp3 does not exist or doesn't return content length")
	}

	if !strings.Contains(err.Error(), "failed to get size") {
		t.Fatalf("Expected error about getting size, got: %v", err)
	}
}
