package commands

import (
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"path/filepath"
	"time"
)

// Iso is a subcommand `podcast-cdr-manager disk iso`
// parent-flag: profile
func Iso(profile string) {
}

// IsoGenerate is a subcommand `podcast-cdr-manager disk iso generate`
// parent-flag: profile
// Flags:
//   dry: --dry (default: true) When this is true it's a dry run and doesn't save changes
//   includeData: --include-data (default: true) Include the profile data for backup sake
//   index: --index (default: -1) The chosen disk index number see disk list
//   outputDirectory: --output-dir (default: ".") Directory to output the ISO
func IsoGenerate(profile string, dry bool, includeData bool, index int, outputDirectory string) error {
	profile = GetProfile(profile)
	p, err := podcast_cdr_manager.OpenProfile(profile)
	if err != nil {
		return fmt.Errorf("opening profile: %w", err)
	}
	disk, err := p.GetDiskByIndex(index)
	if err != nil {
		return fmt.Errorf("finding disk: %w", err)
	}
	if err := disk.GenerateIso(filepath.Join(outputDirectory, disk.Filename), includeData, p); err != nil {
		return fmt.Errorf("generating iso: %w", err)
	}
	now := time.Now()
	disk.BurntDate = &now
	fmt.Printf("Disk marked as burnt\n")
	if !dry {
		if err := p.Save(); err != nil {
			return fmt.Errorf("saving profile: %w", err)
		}
	} else {
		fmt.Printf("Dry not not saving to changes to profile\n")
	}
	fmt.Printf("ISO generated\n")
	return nil
}
