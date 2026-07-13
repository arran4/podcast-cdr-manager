package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillCommandsIntegration(t *testing.T) {
	// Create mock skill source
	srcDir, err := os.MkdirTemp("", "mock-skill-source")
	if err != nil {
		t.Fatalf("Failed to create temp src dir: %v", err)
	}
	defer os.RemoveAll(srcDir)

	if err := os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("# Mock Skill"), 0644); err != nil {
		t.Fatalf("Failed to write mock SKILL.md: %v", err)
	}

	// Make sure we're in a clean state for project scope
	basePath, _ := os.Getwd()
	agentsDir := filepath.Join(basePath, ".agents")
	os.RemoveAll(agentsDir)
	defer os.RemoveAll(agentsDir)

	// Install
	err = SkillInstall("test-profile", "project", "codex", false, srcDir)
	if err != nil {
		t.Fatalf("SkillInstall failed: %v", err)
	}

	skillName := filepath.Base(srcDir)

	// Inspect
	err = SkillInspect("test-profile", "project", "codex", skillName)
	if err != nil {
		t.Fatalf("SkillInspect failed: %v", err)
	}

	// List
	err = SkillList("test-profile", true, "project")
	if err != nil {
		t.Fatalf("SkillList failed: %v", err)
	}

	// Update
	err = SkillUpdate("test-profile", false, false, "project", "codex", skillName)
	if err != nil {
		t.Fatalf("SkillUpdate failed: %v", err)
	}

	// Remove
	err = SkillRemove("test-profile", "project", "codex", skillName)
	if err != nil {
		t.Fatalf("SkillRemove failed: %v", err)
	}
}
