package commands

import (
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"math"
	"slices"
	"strings"
	"time"
)

// DiskNext is a subcommand `podcast-cdr-manager disk next`
// parent-flag: profile
// Flags:
//   dedicatedIndex: --dedicatedIndex (default: -1) Use the podcast index (number at start of column) using this makes the podcast dedicated
//   diskSizeMb: --disk-size-mb (default: 600) The disk size in MB
//   create: --create (default: false) Create disk if not found
//   dry: --dry (default: true) When this is true it's a dry run and doesn't save changes
//   -- If you don't specify a section, this creates the disk, use -dedicatedIndex to produce a disk with
//   -- only one podcast on it, otherwise it will try fill the disk
func DiskNext(profile string, dedicatedIndex int, diskSizeMb int, create bool, dry bool) error {
	profile = GetProfile(profile)
	p, err := podcast_cdr_manager.OpenProfile(profile)
	if err != nil {
		return fmt.Errorf("opening profile: %w", err)
	}
	var sub *podcast_cdr_manager.Subscription
	if dedicatedIndex != -1 {
		sub, err = p.GetSubByIndex(dedicatedIndex)
		if err != nil {
			return fmt.Errorf("getting subscription by index %d: %w", dedicatedIndex, err)
		}
	}
	var disk *podcast_cdr_manager.Disk
	if sub != nil {
		if disk, err = p.FindFreeDiskForSubscription(sub.Url); err != nil {
			return fmt.Errorf("finding disk: %w", err)
		}
	} else {
		if disk, err = p.FindFreeDisk(); err != nil {
			return fmt.Errorf("finding disk: %w", err)
		}
	}
	if disk == nil {
		if !create {
			return fmt.Errorf("disk does not exist")
		}
		subscriptionUrlFilter := []string{}
		if sub != nil {
			subscriptionUrlFilter = append(subscriptionUrlFilter, sub.Url)
		}
		disk, err = p.CreateDisk(subscriptionUrlFilter, diskSizeMb)
		if err != nil {
			return fmt.Errorf("creating disk: %w", err)
		}
	}
	casts, err := p.ListUnassignedCasts()
	if err != nil {
		return fmt.Errorf("getting unassigned podcast episodes: %w", err)
	}
	if sub != nil {
		casts = slices.DeleteFunc(casts, func(cast *podcast_cdr_manager.Cast) bool {
			return !strings.EqualFold(cast.SubscriptionUrl, sub.Url)
		})
	}
	allocations := 0
	for _, cast := range casts {
		castSizeMb := int(math.Ceil(float64(*cast.SizeBytes) / 1024.0 / 1024.0))
		if disk.UsedSpaceMb+castSizeMb >= disk.TotalSpaceMb {
			now := time.Now()
			disk.ReadyToBurn = &now
			fmt.Printf("Ran out of space on %s (%d mb + %d mb / %d mb) (Meaning it's ready to generate an ISO from)\n", disk.Filename, disk.UsedSpaceMb, castSizeMb, disk.TotalSpaceMb)
			break
		}
		fmt.Printf("Put %s on %s (%d mb + %d mb / %d mb)\n", cast.MpegLink, disk.Filename, disk.UsedSpaceMb, castSizeMb, disk.TotalSpaceMb)
		cast.DiskName = disk.Name
		// TODO make it get the size some-other way, until then hard fail.
		disk.UsedSpaceMb += castSizeMb
		allocations++
	}
	if allocations == 0 {
		fmt.Printf("No allocations made, exiting withsout saving\n")
		return nil
	}
	if !dry {
		if err := p.Save(); err != nil {
			return fmt.Errorf("saving profile: %w", err)
		}
	} else {
		fmt.Printf("Dry not not saving to changes to profile\n")
	}
	return nil
}
