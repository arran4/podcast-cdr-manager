package skill

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"path/filepath"
	"strings"
)

type Source interface {
	Fetch(dest string) (revision string, err error)
}

type LocalSource struct {
	Path string
}

func (s *LocalSource) Fetch(dest string) (string, error) {
	st, err := os.Stat(s.Path)
	if err != nil {
		return "", fmt.Errorf("local source error: %w", err)
	}

	if !st.IsDir() {
		return "", fmt.Errorf("local source must be a directory")
	}

	err = copyDir(s.Path, dest)
	if err != nil {
		return "", err
	}

	return "local-" + hashDir(dest), nil
}

type GitHubSource struct {
	Owner string
	Repo  string
	Path  string // Subdirectory within the repo
	Ref   string // branch/tag/commit, empty defaults to main/master
}

type githubCommit struct {
	Sha string `json:"sha"`
}

func (s *GitHubSource) getLatestCommit() (string, error) {
	ref := s.Ref
	if ref == "" {
		ref = "HEAD" // Simplifying for now, might need to resolve HEAD to a branch
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", s.Owner, s.Repo, ref)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	// GitHub recommends a user agent
	req.Header.Set("User-Agent", "podcast-cdr-manager-skill-installer")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch commit info: %s", resp.Status)
	}

	var commit githubCommit
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return "", err
	}

	return commit.Sha, nil
}

func (s *GitHubSource) Fetch(dest string) (string, error) {
	sha, err := s.getLatestCommit()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", s.Owner, s.Repo, sha)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download from github: %s", resp.Status)
	}

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = gzr.Close()
	}()

	tr := tar.NewReader(gzr)

	expectedPrefix := fmt.Sprintf("%s-%s/", s.Repo, sha)
	if s.Path != "" {
		expectedPrefix = filepath.Join(expectedPrefix, s.Path)
		if !strings.HasSuffix(expectedPrefix, "/") {
			expectedPrefix += "/"
		}
	}

	foundSkillMd := false

	// IMPORTANT: Pre-calculate absolute clean paths to avoid traversal
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return "", err
	}
	absDest = filepath.Clean(absDest)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Normalize paths to avoid slash/backslash issues on Windows, though tar is usually /
		name := filepath.ToSlash(header.Name)

		if !strings.HasPrefix(name, expectedPrefix) {
			continue
		}

		// Remove the prefix to get the path relative to our dest
		relPath := strings.TrimPrefix(name, expectedPrefix)
		if relPath == "" {
			continue
		}

		target := filepath.Join(absDest, relPath)

		// Prevent path traversal
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return "", err
		}
		absTarget = filepath.Clean(absTarget)

		if !strings.HasPrefix(absTarget, absDest+string(os.PathSeparator)) && absTarget != absDest {
			return "", fmt.Errorf("invalid path in tarball: %s", target)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return "", err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644) // Restrict to non-executable
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return "", err
			}
			_ = f.Close()
			if relPath == "SKILL.md" {
				foundSkillMd = true
			}
		}
	}

	if !foundSkillMd {
		return "", fmt.Errorf("SKILL.md not found in source")
	}

	return sha, nil
}

func ParseSource(sourceStr string) (Source, error) {
	if strings.HasPrefix(sourceStr, "./") || strings.HasPrefix(sourceStr, "/") || strings.HasPrefix(sourceStr, "../") {
		return &LocalSource{Path: sourceStr}, nil
	}

	// Check if it exists locally as a directory
	if fi, err := os.Stat(sourceStr); err == nil && fi.IsDir() {
		return &LocalSource{Path: sourceStr}, nil
	}

	parts := strings.Split(sourceStr, "/")
	if len(parts) >= 2 {
		s := &GitHubSource{
			Owner: parts[0],
			Repo:  parts[1],
		}
		if len(parts) > 2 {
			s.Path = strings.Join(parts[2:], "/")
		}
		return s, nil
	}

	return nil, fmt.Errorf("invalid source format: %s", sourceStr)
}

func copyDir(src string, dst string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	absSrc = filepath.Clean(absSrc)

	absDest, err := filepath.Abs(dst)
	if err != nil {
		return err
	}
	absDest = filepath.Clean(absDest)

	return filepath.Walk(absSrc, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(absSrc, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(absDest, relPath)
		absTarget, err := filepath.Abs(targetPath)
		if err != nil {
			return err
		}
		absTarget = filepath.Clean(absTarget)

		// Prevent path traversal
		if !strings.HasPrefix(absTarget, absDest+string(os.PathSeparator)) && absTarget != absDest {
			return fmt.Errorf("invalid copy path: %s", targetPath)
		}

		if info.IsDir() {
			return os.MkdirAll(absTarget, 0755)
		}

		err = func() error {
			srcF, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() {
				_ = srcF.Close()
			}()

			dstF, err := os.OpenFile(absTarget, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
			if err != nil {
				return err
			}
			defer func() {
				_ = dstF.Close()
			}()

			_, err = io.Copy(dstF, srcF)
			return err
		}()
		return err
	})
}

func hashDir(dir string) string {
	h := sha256.New()
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".skill-metadata.json") {
			return nil
		}
		func() {
			f, err := os.Open(path)
			if err != nil {
				return
			}
			defer func() {
				_ = f.Close()
			}()
			_, _ = io.Copy(h, f)
		}()
		return nil
	})
	return hex.EncodeToString(h.Sum(nil))[:8]
}
