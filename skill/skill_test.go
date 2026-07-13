package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalSourceFetch(t *testing.T) {
	// Create a temporary mock skill directory
	srcDir, err := os.MkdirTemp("", "mock-skill-src")
	if err != nil {
		t.Fatalf("Failed to create temp src dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	skillMdPath := filepath.Join(srcDir, "SKILL.md")
	if err := os.WriteFile(skillMdPath, []byte("# Test Skill"), 0644); err != nil {
		t.Fatalf("Failed to write mock SKILL.md: %v", err)
	}

	src := &LocalSource{Path: srcDir}

	destDir, err := os.MkdirTemp("", "mock-skill-dest")
	if err != nil {
		t.Fatalf("Failed to create temp dest dir: %v", err)
	}
	defer os.RemoveAll(destDir)

	revision, err := src.Fetch(destDir)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if revision == "" {
		t.Errorf("Expected non-empty revision")
	}

	if _, err := os.Stat(filepath.Join(destDir, "SKILL.md")); os.IsNotExist(err) {
		t.Errorf("SKILL.md was not copied to destination")
	}
}

func TestInstallerLocal(t *testing.T) {
	// Setup local source
	srcDir, err := os.MkdirTemp("", "mock-skill-src")
	if err != nil {
		t.Fatalf("Failed to create temp src dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	skillMdPath := filepath.Join(srcDir, "SKILL.md")
	if err := os.WriteFile(skillMdPath, []byte("# Test Skill"), 0644); err != nil {
		t.Fatalf("Failed to write mock SKILL.md: %v", err)
	}

	// Setup custom target directory by mocking HOME/CWD via env isn't trivial in process,
	// so we'll just test the core target logic directly.
	target, err := NewTarget("common", "project")
	if err != nil {
		t.Fatalf("Failed to create target: %v", err)
	}

	// We override the base path in testing if needed, or just let it use CWD.
	// We'll let it use CWD for project scope.
	skillName := filepath.Base(srcDir)
	dest := target.InstallPath(skillName)
	defer os.RemoveAll(filepath.Dir(filepath.Dir(dest))) // Cleanup .agents in CWD

	inst := NewInstaller()

	// Install
	err = inst.Install(srcDir, target, false)
	if err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	// Check if installed
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Fatalf("Skill directory not created at %s", dest)
	}

	// Update the digest explicitly so subsequent Update check works
	// hashDir depends on the contents, and we didn't add the metadata file properly for testing

	// Check metadata
	md, err := ReadMetadata(dest)
	if err != nil {
		t.Fatalf("ReadMetadata failed: %v", err)
	}
	if md.Source != srcDir {
		t.Errorf("Expected source %s, got %s", srcDir, md.Source)
	}

	// Update (No changes)
	err = inst.Update(skillName, target, false)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Remove
	err = inst.Remove(skillName, target)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("Skill directory still exists after removal")
	}
}

func TestParseSource(t *testing.T) {
	s, err := ParseSource("owner/repo/path/to/skill")
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	gh, ok := s.(*GitHubSource)
	if !ok {
		t.Fatalf("Expected GitHubSource")
	}
	if gh.Owner != "owner" || gh.Repo != "repo" || gh.Path != "path/to/skill" {
		t.Errorf("Incorrect GitHubSource parsing: %+v", gh)
	}

	s, err = ParseSource("./local-skill")
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}
	loc, ok := s.(*LocalSource)
	if !ok {
		t.Fatalf("Expected LocalSource")
	}
	if loc.Path != "./local-skill" {
		t.Errorf("Incorrect LocalSource parsing: %+v", loc)
	}
}
