package podcast_cdr_manager

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDiskfsISOUsesOnlyDeclaredVolumeSectors(t *testing.T) {
	payload := []byte("podcast audio")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{{MpegLink: server.URL + "/podcast.mp3", DiskName: disk.Name}}}
	image := filepath.Join(t.TempDir(), "podcast.iso")
	if err := disk.GenerateIso(image, false, profile); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(image)
	if err != nil {
		t.Fatal(err)
	}
	// The writer provisions >= 4 MiB of sparse image headroom, but the
	// published ISO must contain *only* the sectors described by the PVD.
	if len(contents) < 17*2048 || string(contents[16*2048+1:16*2048+6]) != "CD001" {
		t.Fatal("invalid ISO9660 descriptor")
	}
	sectors := binary.LittleEndian.Uint32(contents[16*2048+80 : 16*2048+84])
	if len(contents) != int(sectors)*2048 {
		t.Fatalf("ISO length %d differs from declared volume size %d", len(contents), sectors*2048)
	}
	if len(contents) >= 1<<20 {
		t.Fatalf("unexpected ISO overhead: %d bytes for %d-byte MP3", len(contents), len(payload))
	}
}

func TestDiskfsJolietAndRockRidgePreserveUnicodeFilename(t *testing.T) {
	if _, err := exec.LookPath("xorriso"); err != nil {
		t.Skip("xorriso is required for independent ISO extraction")
	}
	const filename = "Épisode français 2026.mp3"
	payload := []byte("audio with Unicode filename")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{{
		MpegLink: server.URL + "/%C3%89pisode%20fran%C3%A7ais%202026.mp3",
		DiskName: disk.Name,
	}}}
	work := t.TempDir()
	isoPath := filepath.Join(work, "disc.iso")
	if err := disk.GenerateIso(isoPath, false, profile); err != nil {
		t.Fatalf("generate Unicode ISO: %v", err)
	}
	extracted := filepath.Join(work, "extracted")
	if err := os.Mkdir(extracted, 0755); err != nil { t.Fatal(err) }
	cmd := exec.Command("xorriso", "-osirrox", "on", "-indev", isoPath, "-extract", "/", extracted)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("extract Unicode ISO: %v\n%s", err, output)
	}
	got, err := os.ReadFile(filepath.Join(extracted, filename))
	if err != nil { t.Fatal(err) }
	if !bytes.Equal(got, payload) {
		t.Fatalf("Unicode podcast contents mismatch: %q", got)
	}
}
