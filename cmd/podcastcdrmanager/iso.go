package main

import (
	"flag"
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"os"
	"path/filepath"
	"time"
)

type IsoConfig struct{}

type IsoCliPoint func(remainingArgs []string, mc *MainConfig, dc *DisksConfig, ic *IsoConfig) error

var (
	IsoSections = map[string]IsoCliPoint{
		"help": RunIsoHelp,
		"":     RunIsoHelp,
		"generate": func(remainingArgs []string, mc *MainConfig, dc *DisksConfig, ic *IsoConfig) error {
			fs := flag.NewFlagSet("iso-rss", flag.ExitOnError)
			dry := fs.Bool("dry", true, "When this is true it's a dry run and doesn't save changes")
			includeData := fs.Bool("include-data", true, "Include the profile data for backup sake")
			index := fs.Int("index", -1, "The chosen disk index number see disk list")
			outputDirectory := fs.String("output-dir", ".", "Directory to output the ISO")
			if err := fs.Parse(remainingArgs); err != nil {
				return fmt.Errorf("parsing arguments: %w", err)
			}
			profile, err := podcast_cdr_manager.OpenProfile(mc.profile)
			if err != nil {
				return fmt.Errorf("opening profile: %w", err)
			}
			disk, err := profile.GetDiskByIndex(*index)
			if err != nil {
				return fmt.Errorf("finding disk: %w", err)
			}
			if err := disk.GenerateIso(filepath.Join(*outputDirectory, disk.Filename), *includeData, profile); err != nil {
				return fmt.Errorf("generating iso: %w", err)
			}
			now := time.Now()
			disk.BurntDate = &now
			fmt.Printf("Disk marked as burnt\n")
			if !*dry {
				if err := profile.Save(); err != nil {
					return fmt.Errorf("saving profile: %w", err)
				}
			} else {
				fmt.Printf("Dry not not saving to changes to profile\n")
			}
			fmt.Printf("ISO generated\n")
			return nil
		},
	}
)

func RunIsoHelp(args []string, mc *MainConfig, dc *DisksConfig, ic *IsoConfig) error {
	fmt.Printf("usage: %s [options] [sections] [...]\n", os.Args[0])
	fmt.Printf("\toptions:\n")
	fmt.Printf("%19s %-20s %-39s\n", "-help", "", "This")
	fmt.Printf("\tsections:\n")
	fmt.Printf("%19s %-20s %-39s\n", "help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "generate", "[-dry] -index <index>", "Generates the ISO from provided disk index number (see disk table)")
	return nil
}

func RunIso(remainingArgs []string, mc *MainConfig, dc *DisksConfig) error {
	fs := flag.NewFlagSet("iso", flag.ExitOnError)
	help := fs.Bool("help", false, "")
	if err := fs.Parse(remainingArgs); err != nil {
		return fmt.Errorf("formatting args: %s", err)
	}
	ic := &IsoConfig{}
	if *help {
		if err := RunIsoHelp(fs.Args(), mc, dc, ic); err != nil {
			return fmt.Errorf("running help: %s", err)
		}
	}
	section, ok := IsoSections[fs.Arg(0)]
	if !ok {
		section = RunIsoHelp
		fmt.Printf("Failed to find %s\n", fs.Arg(0))
	}
	if err := section(podcast_cdr_manager.SkipFirstN(fs.Args(), 1), mc, dc, ic); err != nil {
		return fmt.Errorf("running %s: %s", fs.Arg(0), err)
	}
	return nil
}
