package podcast_cdr_manager

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kdomanski/iso9660"
)

func TestGenerateISORoundTrip(t *testing.T) {
	audio := []byte("test mp3 payload")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(audio)
	}))
	defer server.Close()

	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Name: "test-profile", Version: "1", Casts: []*Cast{{
		MpegLink: server.URL + "/episode.mp3", DiskName: disk.Name,
	}}}
	target := filepath.Join(t.TempDir(), "podcasts.iso")
	if err := disk.GenerateIso(target, true, profile); err != nil {
		t.Fatalf("generate: %v", err)
	}
	f, err := os.Open(target)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	image, err := iso9660.OpenImage(f)
	if err != nil {
		t.Fatal(err)
	}
	root, err := image.RootDir()
	if err != nil {
		t.Fatal(err)
	}
	children, err := root.GetChildren()
	if err != nil {
		t.Fatal(err)
	}
	foundMP3, foundYAML := false, false
	for _, child := range children {
		data, err := io.ReadAll(child.Reader())
		if err != nil {
			t.Fatal(err)
		}
		switch strings.ToLower(child.Name()) {
		case "episode.mp3":
			foundMP3 = true
			if !bytes.Equal(data, audio) {
				t.Fatalf("MP3 bytes do not match")
			}
		case "test-profile.yaml":
			foundYAML = true
			if !bytes.Contains(data, []byte("test-profile")) {
				t.Fatalf("profile backup missing name")
			}
		}
	}
	if !foundMP3 || !foundYAML {
		t.Fatalf("missing output entries: MP3=%v YAML=%v", foundMP3, foundYAML)
	}
}

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

func TestGenerateISORejectsMangledNameCollisionBeforeDownload(t *testing.T) {
	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{
		{MpegLink: "https://example.invalid/" + strings.Repeat("a", 40) + ".mp3", DiskName: disk.Name},
		{MpegLink: "https://example.invalid/" + strings.Repeat("a", 39) + "b.mp3", DiskName: disk.Name},
	}}
	err := disk.GenerateIso(filepath.Join(t.TempDir(), "out.iso"), false, profile)
	if err == nil || !strings.Contains(err.Error(), "collision") {
		t.Fatalf("expected ISO filename collision, got %v", err)
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
