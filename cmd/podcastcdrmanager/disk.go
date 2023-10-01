package main

import (
	"flag"
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"os"
)

type DisksConfig struct{}

type DisksCliPoint func(remainingArgs []string, mc *MainConfig, sc *DisksConfig) error

var (
	DisksSections = map[string]DisksCliPoint{
		"help": RunDiskHelp,
		"":     RunDiskHelp,
		"list": func(remainingArgs []string, mc *MainConfig, sc *DisksConfig) error {
			profile, err := podcast_cdr_manager.OpenProfile(mc.profile)
			if err != nil {
				return fmt.Errorf("opening profile: %w", err)
			}
			disks, err := profile.ListDisks()
			if err != nil {
				return fmt.Errorf("getting disks: %w", err)
			}
			fmt.Printf("    %30s %30s %s\n", "Disk filename", "Created Date", "Burnt Date")
			for i, disk := range disks {
				fmt.Printf("% 3d %30s %30s %s\n", i, fmt.Sprint(disk.CreatedDate), fmt.Sprint(disk.BurntDate), disk.Filename)
			}
			return nil
		},
	}
)

func RunDiskHelp(args []string, mc *MainConfig, sc *DisksConfig) error {
	fmt.Printf("usage: %s [options] [sections] [...]\n", os.Args[0])
	fmt.Printf("\toptions:\n")
	fmt.Printf("%19s %-20s %-39s\n", "-help", "", "This")
	fmt.Printf("\tsections:\n")
	fmt.Printf("%19s %-20s %-39s\n", "help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "list", "", "List disks")
	return nil
}

func RunDisk(remainingArgs []string, mc *MainConfig) error {
	fs := flag.NewFlagSet("disks", flag.ExitOnError)
	help := fs.Bool("help", false, "")
	if err := fs.Parse(remainingArgs); err != nil {
		return fmt.Errorf("formatting args: %s", err)
	}
	sc := &DisksConfig{}
	if *help {
		if err := RunDiskHelp(fs.Args(), mc, sc); err != nil {
			return fmt.Errorf("running help: %s", err)
		}
	}
	section, ok := DisksSections[fs.Arg(1)]
	if !ok {
		section = RunDiskHelp
		fmt.Printf("Failed to find %s\n", fs.Arg(1))
	}
	if err := section(append([]string{fs.Arg(0)}, fs.Args()[min(2, len(fs.Args())):]...), mc, sc); err != nil {
		return fmt.Errorf("running help: %s", err)
	}
	return nil
}
