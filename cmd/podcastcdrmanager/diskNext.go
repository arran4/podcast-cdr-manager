package main

import (
	"flag"
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"math"
	"os"
	"slices"
	"strings"
	"time"
)

type DiskNextConfig struct{}

type DiskNextCliPoint func(remainingArgs []string, mc *MainConfig, sc *DiskNextConfig) error

var (
	DiskNextSections = map[string]DiskNextCliPoint{
		"help": RunDiskNextHelp,
	}
)

func RunDiskNextHelp(args []string, mc *MainConfig, sc *DiskNextConfig) error {
	fmt.Printf("usage: %s [options] [sections] [...]\n", os.Args[0])
	fmt.Printf("\toptions:\n")
	fmt.Printf("%19s %-20s %-39s\n", "-help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "-dedicatedIndex", "<number:-1>", "Use the podcast index (number at start of column) using this makes the podcast dedicated")
	fmt.Printf("%19s %-20s %-39s\n", "-dry", "<bool:true>", "When true doesn't save changes -- preview mode")
	fmt.Printf("\tsections:\n")
	fmt.Printf("%19s %-20s %-39s\n", "help", "", "This")
	fmt.Printf("\n")
	fmt.Printf("If you don't specify a section, this creates the disk, use -dedicatedIndex to produce a disk with\n")
	fmt.Printf("only one podcast on it, otherwise it will try fill the disk\n")
	fmt.Printf("\n")
	return nil
}

func RunDiskNext(remainingArgs []string, mc *MainConfig, dc *DisksConfig) error {
	fs := flag.NewFlagSet("disk-next", flag.ExitOnError)
	help := fs.Bool("help", false, "")
	dedicatedIndex := fs.Int("dedicatedIndex", -1, "Use the podcast index (number at start of column) using this makes the podcast dedicated")
	diskSizeMb := fs.Int("disk-size-mb", 600, "The disk size in MB")
	create := fs.Bool("create", false, "Create disk if not found")
	dry := fs.Bool("dry", true, "When this is true it's a dry run and doesn't save changes")
	if err := fs.Parse(remainingArgs); err != nil {
		return fmt.Errorf("formatting args: %s", err)
	}
	return DoRunDiskNext(help, fs, mc, dedicatedIndex, create, diskSizeMb, dry)
}

func DoRunDiskNext(help *bool, fs *flag.FlagSet, mc *MainConfig, dedicatedIndex *int, create *bool, diskSizeMb *int, dry *bool) error {
	sc := &DiskNextConfig{}
	if *help {
		if err := RunDiskNextHelp(fs.Args(), mc, sc); err != nil {
			return fmt.Errorf("running help: %s", err)
		}
	}
	section, ok := DiskNextSections[fs.Arg(0)]
	if ok {
		if err := section(podcast_cdr_manager.SkipFirstN(fs.Args(), 1), mc, sc); err != nil {
			return fmt.Errorf("running help: %s", err)
		}
	}
	profile, err := podcast_cdr_manager.OpenProfile(mc.profile)
	if err != nil {
		return fmt.Errorf("opening profile: %w", err)
	}
	var sub *podcast_cdr_manager.Subscription
	if *dedicatedIndex != -1 {
		sub, err = profile.GetSubByIndex(*dedicatedIndex)
		if err != nil {
			return fmt.Errorf("getting subscription by index %d: %w", *dedicatedIndex, err)
		}
	}
	var disk *podcast_cdr_manager.Disk
	if sub != nil {
		if disk, err = profile.FindFreeDiskForSubscription(sub.Url); err != nil {
			return fmt.Errorf("finding disk: %w", err)
		}
	} else {
		if disk, err = profile.FindFreeDisk(); err != nil {
			return fmt.Errorf("finding disk: %w", err)
		}
	}
	if disk == nil {
		if !*create {
			return fmt.Errorf("disk does not exist")
		}
		subscriptionUrlFilter := []string{}
		if sub != nil {
			subscriptionUrlFilter = append(subscriptionUrlFilter, sub.Url)
		}
		disk, err = profile.CreateDisk(subscriptionUrlFilter, *diskSizeMb)
		if err != nil {
			return fmt.Errorf("creating disk: %w", err)
		}
	}
	casts, err := profile.ListUnassignedCasts()
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
		// TODO: make it get the size some-other way, until then hard fail.
		disk.UsedSpaceMb += castSizeMb
		allocations++
	}
	if allocations == 0 {
		fmt.Printf("No allocations made, exiting withsout saving\n")
		return nil
	}
	if !*dry {
		if err := profile.Save(); err != nil {
			return fmt.Errorf("saving profile: %w", err)
		}
	} else {
		fmt.Printf("Dry not not saving to changes to profile\n")
	}
	return nil
}
