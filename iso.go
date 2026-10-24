package podcast_cdr_manager

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type podcastISOInput struct {
	source string
	target string
}

// legacyISOIdentifier matches the current writer's file-name mangling. Reject
// collisions before AddLocalFile (which otherwise replaces staged files).
// Replace this with backend-specific collision handling when migrating writers.
func legacyISOIdentifier(name string) string {
	parts := strings.Split(strings.ToLower(name), ".")
	base := parts[0]
	extension := ""
	if len(parts) > 1 {
		base = strings.Join(parts[:len(parts)-1], "_")
		extension = parts[len(parts)-1]
	}
	mangle := func(s string, max int) string {
		const allowed = "abcdefghijklmnopqrstuvwxyz0123456789_!\"%&'()*+,-./:;<=>?"
		var out strings.Builder
		for i := 0; i < len(s) && i < max; i++ {
			if strings.ContainsRune(allowed, rune(s[i])) {
				out.WriteByte(s[i])
			} else {
				out.WriteByte('_')
			}
		}
		return out.String()
	}
	extension = mangle(extension, 8)
	maxBase := 28 // identifier length 30 less ";1"
	if extension != "" {
		maxBase -= len(extension) + 1
	}
	base = mangle(base, maxBase)
	if extension != "" {
		return base + "." + extension + ";1"
	}
	return base + ";1"
}

func podcastFilename(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", fmt.Errorf("invalid podcast URL %q: %w", link, err)
	}
	if (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("unsupported podcast URL %q", link)
	}
	name := path.Base(u.Path)
	if name == "." || name == "/" || name == "" {
		return "", fmt.Errorf("podcast URL has no filename: %q", link)
	}
	return name, nil
}

func downloadPodcast(ctx context.Context, link, destination string, maxBytes int64) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download %q: %w", link, err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("download %q: unexpected HTTP status %s", link, response.Status)
	}
	if maxBytes > 0 && response.ContentLength > maxBytes {
		return fmt.Errorf("download %q: exceeds disc capacity", link)
	}

	out, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create download file: %w", err)
	}
	defer out.Close()

	var reader io.Reader = response.Body
	if maxBytes > 0 {
		reader = io.LimitReader(response.Body, maxBytes+1)
	}
	n, copyErr := io.Copy(out, reader)
	if copyErr != nil {
		return fmt.Errorf("download %q: %w", link, copyErr)
	}
	if maxBytes > 0 && n > maxBytes {
		return fmt.Errorf("download %q: exceeds disc capacity", link)
	}
	if response.ContentLength >= 0 && n != response.ContentLength {
		return fmt.Errorf("download %q: expected %d bytes, received %d", link, response.ContentLength, n)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("finish download %q: %w", link, err)
	}
	return nil
}

func fetchPodcasts(ctx context.Context, casts []*Cast, filenames []string, stagingDir string, maxBytes int64) ([]podcastISOInput, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	inputs := make([]podcastISOInput, len(casts))
	jobs := make(chan int)
	var workers sync.WaitGroup
	var errorMu sync.Mutex
	var firstErr error

	worker := func() {
		defer workers.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case i, ok := <-jobs:
				if !ok {
					return
				}
				localFile := filepath.Join(stagingDir, fmt.Sprintf("episode-%06d", i))
				if err := downloadPodcast(ctx, casts[i].MpegLink, localFile, maxBytes); err != nil {
					errorMu.Lock()
					if firstErr == nil {
						firstErr = err
						cancel()
					}
					errorMu.Unlock()
					return
				}
				inputs[i] = podcastISOInput{source: localFile, target: filenames[i]}
			}
		}
	}
	for i := 0; i < len(casts) && i < 4; i++ {
		workers.Add(1)
		go worker()
	}
submit:
	for i := range casts {
		select {
		case <-ctx.Done():
			break submit
		case jobs <- i:
		}
	}
	close(jobs)
	workers.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return inputs, nil
}

// GenerateIso downloads each episode into temporary files, then produces a
// complete image before replacing the destination. Download memory is bounded.
func (d *Disk) GenerateIso(fn string, includeProfileData bool, profile *Profile) error {
	return d.generateISO(context.Background(), fn, includeProfileData, profile, newISOImageWriter)
}

func (d *Disk) generateISO(ctx context.Context, fn string, includeProfileData bool, profile *Profile, createWriter func() (isoImageWriter, error)) error {
	if d == nil || profile == nil {
		return fmt.Errorf("disk and profile must not be nil")
	}
	casts, err := profile.GetCastsByDiskName(d.Name)
	if err != nil {
		return fmt.Errorf("find casts by disk: %w", err)
	}
	if len(casts) == 0 {
		return fmt.Errorf("found 0 casts for disk")
	}
	names := make([]string, len(casts))
	seen := make(map[string]string)
	if includeProfileData {
		seen[legacyISOIdentifier(profile.Name+".yaml")] = profile.Name + ".yaml"
	}
	for i, cast := range casts {
		if cast == nil {
			return fmt.Errorf("nil podcast at index %d", i)
		}
		name, err := podcastFilename(cast.MpegLink)
		if err != nil {
			return err
		}
		key := legacyISOIdentifier(name)
		if previous, exists := seen[key]; exists {
			return fmt.Errorf("ISO filename collision: %q and %q both become %q", previous, name, key)
		}
		seen[key] = name
		names[i] = name
	}

	stagingDir, err := os.MkdirTemp("", "podcast-cdr-downloads-")
	if err != nil {
		return fmt.Errorf("create temporary download directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	var capacityBytes int64
	if d.TotalSpaceMb > 0 {
		capacityBytes = int64(d.TotalSpaceMb) * 1024 * 1024
	}
	inputs, err := fetchPodcasts(ctx, casts, names, stagingDir, capacityBytes)
	if err != nil {
		return fmt.Errorf("fetch podcasts: %w", err)
	}

	writer, err := createWriter()
	if err != nil {
		return fmt.Errorf("create ISO writer: %w", err)
	}
	defer func() {
		if err := writer.Cleanup(); err != nil {
			log.Printf("ISO writer cleanup: %v", err)
		}
	}()

	if includeProfileData {
		data, err := profile.ProfileData()
		if err != nil {
			return fmt.Errorf("serialize profile: %w", err)
		}
		if err := writer.AddFile(bytes.NewReader(data), profile.Name+".yaml"); err != nil {
			return fmt.Errorf("add profile: %w", err)
		}
	}
	for _, input := range inputs {
		if err := writer.AddLocalFile(input.source, input.target); err != nil {
			return fmt.Errorf("add podcast %q: %w", input.target, err)
		}
	}

	out, err := os.CreateTemp(filepath.Dir(fn), "."+filepath.Base(fn)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary ISO: %w", err)
	}
	defer os.Remove(out.Name())

	if err := writer.WriteTo(out, d.Name); err != nil {
		out.Close()
		return fmt.Errorf("write ISO image: %w", err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close ISO image: %w", err)
	}
	stat, err := os.Stat(out.Name())
	if err != nil {
		return fmt.Errorf("stat ISO image: %w", err)
	}
	if capacityBytes > 0 && stat.Size() > capacityBytes {
		return fmt.Errorf("ISO image %d bytes exceeds disc capacity %d bytes", stat.Size(), capacityBytes)
	}
	if err := os.Chmod(out.Name(), 0644); err != nil {
		return fmt.Errorf("set ISO image permissions: %w", err)
	}
	if err := os.Rename(out.Name(), fn); err != nil {
		return fmt.Errorf("publish ISO image: %w", err)
	}
	return nil
}
