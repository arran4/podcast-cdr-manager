package podcast_cdr_manager

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The optional xorriso backend must preserve long names and MP3 data without
// introducing any new Go-module dependencies.
func TestXorrisoBackendPreservesLongFilename(t *testing.T) {
	if _, err := exec.LookPath("xorriso"); err != nil {
		t.Skip("xorriso not installed (CI installs it)")
	}
	const filename = "A very long podcast episode name with spaces.mp3"
	audio := []byte("test audio payload")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(audio)
	}))
	defer server.Close()

	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{{
		MpegLink: server.URL + "/A%20very%20long%20podcast%20episode%20name%20with%20spaces.mp3",
		DiskName: disk.Name,
	}}}
	tmp := t.TempDir()
	isoPath := filepath.Join(tmp, "podcasts.iso")
	if err := disk.generateISO(t.Context(), isoPath, false, profile, newXorrisoImageWriter); err != nil {
		t.Fatalf("create xorriso ISO: %v", err)
	}

	extracted := filepath.Join(tmp, "extracted")
	if err := os.Mkdir(extracted, 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("xorriso", "-osirrox", "on", "-indev", isoPath, "-extract", "/", extracted)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract with xorriso: %v\n%s", err, output)
	}
	got, err := os.ReadFile(filepath.Join(extracted, filename))
	if err != nil {
		t.Fatalf("long original filename not preserved: %v", err)
	}
	if !bytes.Equal(got, audio) {
		t.Fatal("extracted MP3 does not match input")
	}
}

func TestXorrisoBackendRetainsDistinctLongNames(t *testing.T) {
	if _, err := exec.LookPath("xorriso"); err != nil {
		t.Skip("xorriso not installed (CI installs it)")
	}
	prefix := strings.Repeat("a", 35)
	first := prefix + "-one.mp3"
	second := prefix + "-two.mp3"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(filepath.Base(r.URL.Path)))
	}))
	defer server.Close()

	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{
		{MpegLink: server.URL + "/" + first, DiskName: disk.Name},
		{MpegLink: server.URL + "/" + second, DiskName: disk.Name},
	}}
	working := t.TempDir()
	image := filepath.Join(working, "two-long-names.iso")
	if err := disk.generateISO(t.Context(), image, false, profile, newXorrisoImageWriter); err != nil {
		t.Fatalf("xorriso should preserve non-colliding long filenames: %v", err)
	}
	extracted := filepath.Join(working, "extracted")
	if err := os.Mkdir(extracted, 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("xorriso", "-osirrox", "on", "-indev", image, "-extract", "/", extracted)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract with xorriso: %v\n%s", err, output)
	}
	for _, name := range []string{first, second} {
		got, err := os.ReadFile(filepath.Join(extracted, name))
		if err != nil {
			t.Fatalf("missing independently named file %q: %v", name, err)
		}
		if string(got) != name {
			t.Fatalf("bad content in %q: %q", name, got)
		}
	}
}
