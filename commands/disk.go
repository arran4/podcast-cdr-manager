package commands

import (
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
)

// Disk is a subcommand `podcast-cdr-manager disk`
// parent-flag: profile
func Disk(profile string) {
}

// DiskList is a subcommand `podcast-cdr-manager disk list-disks`
// parent-flag: profile
func DiskList(profile string) error {
	profile = GetProfile(profile)
	p, err := podcast_cdr_manager.OpenProfile(profile)
	if err != nil {
		return fmt.Errorf("opening profile: %w", err)
	}
	disks, err := p.ListDisks()
	if err != nil {
		return fmt.Errorf("getting disks: %w", err)
	}
	fmt.Printf("    %30s %30s %s\n", "Disk filename", "Created Date", "Burnt Date")
	for i, disk := range disks {
		fmt.Printf("% 3d %30s %30s %s\n", i, fmt.Sprint(disk.CreatedDate), fmt.Sprint(disk.BurntDate), disk.Filename)
	}
	return nil
}

// DiskCreate is a subcommand `podcast-cdr-manager disk create`
// parent-flag: profile
// Flags:
//   diskSizeMb: --disk-size-mb (default: 600) The disk size in MB
func DiskCreate(profile string, diskSizeMb int) error {
	profile = GetProfile(profile)
	p, err := podcast_cdr_manager.OpenProfile(profile)
	if err != nil {
		return fmt.Errorf("opening profile: %w", err)
	}
	disk, err := p.CreateDisk([]string{}, diskSizeMb)
	if err != nil {
		return fmt.Errorf("create disk: %w", err)
	}
	if err := p.Save(); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}
	fmt.Printf("    %30s %30s %s\n", "Disk filename", "Created Date", "Burnt Date")
	fmt.Printf("%3s %30s %30s %s\n", "_", fmt.Sprint(disk.CreatedDate), fmt.Sprint(disk.BurntDate), disk.Filename)
	return nil
}
