package commands

import (
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
)

// Profile is a subcommand `podcast-cdr-manager profile`
// parent-flag: profile
func Profile(profile string) {
}

// ProfileNew is a subcommand `podcast-cdr-manager profile new`
// parent-flag: profile
// Flags:
//   force: --force (default: false) Forcefully overwrite
//   name: @1 The profile name
func ProfileNew(profile string, force bool, name string) error {
	p, err := podcast_cdr_manager.NewProfile(name, force)
	if err != nil {
		return fmt.Errorf("creating profile %s: %w", name, err)
	}
	if err := p.Save(); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}
	fmt.Printf("Profiled\n")
	return nil
}
