package main

import (
	"flag"
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"os"
)

type ProfileConfig struct{}

type ProfileCliPoint func(remainingArgs []string, mc *MainConfig, sc *ProfileConfig) error

var (
	ProfileSections = map[string]ProfileCliPoint{
		"help": RunProfileHelp,
		"":     RunProfileHelp,
		"new": func(remainingArgs []string, mc *MainConfig, sc *ProfileConfig) error {
			fs := flag.NewFlagSet("profile-new", flag.ExitOnError)
			force := fs.Bool("force", false, "Forcefully overwrite")
			if err := fs.Parse(remainingArgs); err != nil {
				return fmt.Errorf("formatting args: %s", err)
			}
			fmt.Printf("%#v\n", fs.Args())
			if *force {
				fmt.Printf("Force turned on\n")
			}
			profile, err := podcast_cdr_manager.NewProfile(fs.Arg(0), *force)
			if err != nil {
				return fmt.Errorf("creating profile %s: %w", fs.Arg(0), err)
			}
			if err := profile.Save(); err != nil {
				return fmt.Errorf("saving profile: %w", err)
			}
			fmt.Printf("Profiled\n")
			return nil
		},
	}
)

func RunProfileHelp(args []string, mc *MainConfig, sc *ProfileConfig) error {
	fmt.Printf("usage: %s [options] [sections] [...]\n", os.Args[0])
	fmt.Printf("\toptions:\n")
	fmt.Printf("%19s %-20s %-39s\n", "-help", "", "This")
	fmt.Printf("\tsections:\n")
	fmt.Printf("%19s %-20s %-39s\n", "help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "new", "[-force:false] <Profile name>", "Creates a new profile")
	return nil
}

func RunProfile(remainingArgs []string, mc *MainConfig) error {
	fs := flag.NewFlagSet("profile", flag.ExitOnError)
	help := fs.Bool("help", false, "")
	if err := fs.Parse(remainingArgs); err != nil {
		return fmt.Errorf("formatting args: %s", err)
	}
	sc := &ProfileConfig{}
	if *help {
		if err := RunProfileHelp(fs.Args(), mc, sc); err != nil {
			return fmt.Errorf("running help: %s", err)
		}
	}
	section, ok := ProfileSections[fs.Arg(0)]
	if !ok {
		section = RunProfileHelp
		fmt.Printf("Failed to find %s\n", fs.Arg(0))
	}
	if err := section(SkipFirstN(fs.Args(), 1), mc, sc); err != nil {
		return fmt.Errorf("running help: %s", err)
	}
	return nil
}
