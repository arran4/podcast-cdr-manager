package skill

import (
	"fmt"
	"os"
	"path/filepath"
)

// Target defines where a skill should be installed.
type Target interface {
	InstallPath(name string) string
	Name() string
	Scope() string
}

type AgentTarget struct {
	agent string
	scope string
	base  string
}

func NewTarget(agent, scope string) (Target, error) {
	if scope == "" {
		scope = "user"
	}
	if agent == "" {
		agent = "common"
	}

	if scope != "user" && scope != "project" {
		return nil, fmt.Errorf("invalid scope: %s. Must be 'user' or 'project'", scope)
	}

	var base string
	var err error
	if scope == "project" {
		base, err = os.Getwd()
	} else {
		base, err = os.UserHomeDir()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base directory: %w", err)
	}

	return &AgentTarget{agent: agent, scope: scope, base: base}, nil
}

func (t *AgentTarget) Name() string {
	return t.agent
}

func (t *AgentTarget) Scope() string {
	return t.scope
}

func (t *AgentTarget) InstallPath(skillName string) string {
	switch t.agent {
	case "codex", "claude", "copilot", "common":
		return filepath.Join(t.base, ".agents", "skills", skillName)
	case "cursor":
		return filepath.Join(t.base, ".cursor", "skills", skillName)
	default:
		return filepath.Join(t.base, ".agents", t.agent, "skills", skillName)
	}
}

// FindAllTargetPaths looks up all standard target directories to aid listing.
func FindAllTargetPaths(scope string) []string {
	var base string
	if scope == "project" {
		base, _ = os.Getwd()
	} else {
		base, _ = os.UserHomeDir()
	}
	return []string{
		filepath.Join(base, ".agents", "skills"),
		filepath.Join(base, ".cursor", "skills"),
	}
}
