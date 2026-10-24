package podcast_cdr_manager

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Independent validation: xorriso checks the image produced by go-diskfs.
func TestGeneratedISOIsReadableByXorriso(t *testing.T) {
	if _, err := exec.LookPath("xorriso"); err != nil {
		t.Skip("xorriso not installed (CI installs it)")
	}
	payload := []byte("podcast-fixture-audio-data")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Name: "testprofile", Version: "1", Casts: []*Cast{{
		MpegLink: server.URL + "/episode.mp3", DiskName: disk.Name,
	}}}
	temp := t.TempDir()
	image := filepath.Join(temp, "podcasts.iso")
	if err := disk.GenerateIso(image, true, profile); err != nil {
		t.Fatalf("generate ISO: %v", err)
	}

	outdir := filepath.Join(temp, "extracted")
	if err := os.Mkdir(outdir, 0755); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("xorriso", "-osirrox", "on", "-indev", image, "-extract", "/", outdir)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("independent xorriso extraction failed: %v\n%s", err, output)
	}
	got, err := os.ReadFile(filepath.Join(outdir, "episode.mp3"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("MP3 payload mismatch: got %q, expected %q", got, payload)
	}
	backup, err := os.ReadFile(filepath.Join(outdir, "testprofile.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(backup, []byte("testprofile")) {
		t.Fatal("extracted profile does not contain profile name")
	}
}

func TestGenerateISOCancelsOtherDownloads(t *testing.T) {
	started := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/slow.mp3":
			close(started)
			<-r.Context().Done()
		case "/fail.mp3":
			<-started
			http.Error(w, "upstream error", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{
		{MpegLink: server.URL + "/slow.mp3", DiskName: disk.Name},
		{MpegLink: server.URL + "/fail.mp3", DiskName: disk.Name},
	}}
	result := make(chan error, 1)
	go func() {
		result <- disk.GenerateIso(filepath.Join(t.TempDir(), "podcasts.iso"), false, profile)
	}()

	select {
	case err := <-result:
		if err == nil || !strings.Contains(err.Error(), "500") {
			t.Fatalf("expected HTTP 500 rather than cancellation, got %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("generation hung after one HTTP request failed")
	}
}

func TestTruncatedPodcastResponseIsRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000")
		_, _ = w.Write([]byte("truncated"))
	}))
	defer server.Close()

	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{{
		MpegLink: server.URL + "/incomplete.mp3", DiskName: disk.Name,
	}}}
	fn := filepath.Join(t.TempDir(), "podcasts.iso")
	if err := disk.GenerateIso(fn, false, profile); err == nil {
		t.Fatal("expected truncated HTTP payload to fail")
	}
	if _, err := os.Stat(fn); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("partial final ISO should not exist; stat: %v", err)
	}
}

func TestCanceledContextNeverPublishesISO(t *testing.T) {
	disk := &Disk{Name: "PODTEST", TotalSpaceMb: 10}
	profile := &Profile{Casts: []*Cast{{
		MpegLink: "https://example.invalid/episode.mp3", DiskName: disk.Name,
	}}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fn := filepath.Join(t.TempDir(), "podcasts.iso")
	if err := disk.generateISO(ctx, fn, false, profile, newISOImageWriter); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if _, err := os.Stat(fn); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("canceled ISO should not exist; stat: %v", err)
	}
}
