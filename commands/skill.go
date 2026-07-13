package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arran4/podcast-cdr-manager/skill"
)

// Skill is a subcommand `podcast-cdr-manager skill`
// parent-flag: profile
func Skill(profile string) {
	fmt.Println("Skill management commands:")
	fmt.Println("  install <source>   Install an agent skill")
	fmt.Println("  update <name>      Update an installed skill")
	fmt.Println("  remove <name>      Remove an installed skill")
	fmt.Println("  list               List installed skills")
	fmt.Println("  inspect <name>     Inspect an installed skill")
}

// SkillInstall is a subcommand `podcast-cdr-manager skill install`
// parent-flag: profile
// Flags:
//	scope: --scope (default: "user") Scope of installation (user or project)
//	agent: --agent (default: "") Target agent (codex, claude, copilot, cursor)
//	force: --force (default: false) Force update, replacing local modifications
// Parameters:
//  source: @1 The source of the skill
func SkillInstall(profile string, scope string, agent string, force bool, source string) error {
	if source == "" {
		return fmt.Errorf("missing skill source")
	}

	t, err := skill.NewTarget(agent, scope)
	if err != nil {
		return err
	}

	inst := skill.NewInstaller()
	return inst.Install(source, t, force)
}

// SkillUpdate is a subcommand `podcast-cdr-manager skill update`
// parent-flag: profile
// Flags:
//	all: --all (default: false) Update all skills
//	force: --force (default: false) Force update, replacing local modifications
//	scope: --scope (default: "user") Scope of installation (user or project)
//	agent: --agent (default: "") Target agent (codex, claude, copilot, cursor)
// Parameters:
//  name: @1 The name of the skill to update
func SkillUpdate(profile string, all bool, force bool, scope string, agent string, name string) error {
	t, err := skill.NewTarget(agent, scope)
	if err != nil {
		return err
	}

	inst := skill.NewInstaller()

	if all {
		// Update all skills logic (requires scanning directories)
		basePaths := skill.FindAllTargetPaths(scope)
		updatedAny := false
		for _, basePath := range basePaths {
			entries, err := os.ReadDir(basePath)
			if err != nil {
				continue
			}
			agentForPath := agent
			if strings.Contains(basePath, ".cursor") {
				agentForPath = "cursor"
			} else if strings.Contains(basePath, ".agents") {
				agentForPath = "common"
			}
			tForPath, err := skill.NewTarget(agentForPath, scope)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if entry.IsDir() {
					err := inst.Update(entry.Name(), tForPath, force)
					if err != nil {
						fmt.Printf("Warning: failed to update %s: %v\n", entry.Name(), err)
					} else {
						updatedAny = true
					}
				}
			}
		}
		if !updatedAny {
			fmt.Println("No skills found to update.")
		}
		return nil
	}

	if name == "" {
		return fmt.Errorf("missing skill name")
	}
	return inst.Update(name, t, force)
}

// SkillRemove is a subcommand `podcast-cdr-manager skill remove`
// parent-flag: profile
// Flags:
//	scope: --scope (default: "user") Scope of installation (user or project)
//	agent: --agent (default: "") Target agent (codex, claude, copilot, cursor)
// Parameters:
//  name: @1 The name of the skill to remove
func SkillRemove(profile string, scope string, agent string, name string) error {
	if name == "" {
		return fmt.Errorf("missing skill name")
	}

	t, err := skill.NewTarget(agent, scope)
	if err != nil {
		return err
	}

	inst := skill.NewInstaller()
	return inst.Remove(name, t)
}

// SkillList is a subcommand `podcast-cdr-manager skill list`
// parent-flag: profile
// Flags:
//	jsonOut: --json (default: false) Output in JSON format
//	scope: --scope (default: "user") Scope of installation (user or project)
func SkillList(profile string, jsonOut bool, scope string) error {
	type SkillInfo struct {
		Name     string `json:"name"`
		Path     string `json:"path"`
		Source   string `json:"source"`
		Revision string `json:"revision"`
	}

	var skills []SkillInfo
	basePaths := skill.FindAllTargetPaths(scope)

	for _, basePath := range basePaths {
		entries, err := os.ReadDir(basePath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillDir := filepath.Join(basePath, entry.Name())
			md, err := skill.ReadMetadata(skillDir)
			if err != nil {
				continue
			}
			skills = append(skills, SkillInfo{
				Name:     entry.Name(),
				Path:     skillDir,
				Source:   md.Source,
				Revision: md.Revision,
			})
		}
	}

	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(skills)
	}

	fmt.Printf("%-20s %-40s %-20s\n", "NAME", "SOURCE", "REVISION")
	for _, s := range skills {
		fmt.Printf("%-20s %-40s %-20s\n", s.Name, s.Source, s.Revision)
	}
	return nil
}

// SkillInspect is a subcommand `podcast-cdr-manager skill inspect`
// parent-flag: profile
// Flags:
//	scope: --scope (default: "user") Scope of installation (user or project)
//	agent: --agent (default: "") Target agent (codex, claude, copilot, cursor)
// Parameters:
//  name: @1 The name of the skill to inspect
func SkillInspect(profile string, scope string, agent string, name string) error {
	if name == "" {
		return fmt.Errorf("missing skill name")
	}

	t, err := skill.NewTarget(agent, scope)
	if err != nil {
		return err
	}

	skillDir := t.InstallPath(name)
	md, err := skill.ReadMetadata(skillDir)
	if err != nil {
		return fmt.Errorf("could not read metadata for skill '%s': %w", name, err)
	}

	fmt.Printf("Skill: %s\n", name)
	fmt.Printf("Path: %s\n", skillDir)
	fmt.Printf("Source: %s\n", md.Source)
	fmt.Printf("Revision: %s\n", md.Revision)
	fmt.Printf("Install Time: %s\n", md.InstallTime.Format("2006-01-02 15:04:05"))

	skillMdPath := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillMdPath); err == nil {
		fmt.Printf("\nSKILL.md found.\n")
	} else {
		fmt.Printf("\nWarning: SKILL.md not found in installation path.\n")
	}

	return nil
}
