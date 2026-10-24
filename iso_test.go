package podcast_cdr_manager

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

)

func TestGenerateISORejectsHTTPStatusAndPreservesImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()
	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{{MpegLink: server.URL + "/missing.mp3", DiskName: disk.Name}}}
	target := filepath.Join(t.TempDir(), "existing.iso")
	if err := os.WriteFile(target, []byte("good previous image"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := disk.GenerateIso(target, false, profile); err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected HTTP 404 error, got %v", err)
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != "good previous image" {
		t.Fatal("existing image was modified")
	}
}

func TestGenerateISORejectsCaseFoldedCollisionBeforeDownload(t *testing.T) {
	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{
		{MpegLink: "https://example.invalid/EPISODE.mp3", DiskName: disk.Name},
		{MpegLink: "https://example.invalid/episode.mp3", DiskName: disk.Name},
	}}
	err := disk.GenerateIso(filepath.Join(t.TempDir(), "out.iso"), false, profile)
	if err == nil || !strings.Contains(err.Error(), "collision") {
		t.Fatalf("expected case-insensitive collision, got %v", err)
	}
}

func TestGenerateISORejectsNilProfile(t *testing.T) {
	err := (&Disk{Name: "PODTEST"}).GenerateIso("ignored.iso", false, nil)
	if err == nil {
		t.Fatal("expected nil profile error")
	}
}

type failingISOImageWriter struct{}

func (failingISOImageWriter) AddFile(io.Reader, string) error { return nil }
func (failingISOImageWriter) AddLocalFile(string, string) error { return nil }
func (failingISOImageWriter) WriteTo(w io.Writer, _ string) error {
	_, _ = w.Write([]byte("incomplete"))
	return io.ErrUnexpectedEOF
}
func (failingISOImageWriter) Cleanup() error { return nil }

func TestGenerateISOWriteFailurePreservesImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("mp3"))
	}))
	defer server.Close()
	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{{MpegLink: server.URL + "/episode.mp3", DiskName: disk.Name}}}
	target := filepath.Join(t.TempDir(), "old.iso")
	if err := os.WriteFile(target, []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	err := disk.generateISO(t.Context(), target, false, profile, func() (isoImageWriter, error) {
		return failingISOImageWriter{}, nil
	})
	if err == nil || !strings.Contains(err.Error(), "unexpected EOF") {
		t.Fatalf("expected write failure, got %v", err)
	}
	after, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != "old" {
		t.Fatal("failed ISO replaced existing image")
	}
}
