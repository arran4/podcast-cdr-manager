package main

import (
	"flag"
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"os"
)

type CastsConfig struct{}

type CastsCliPoint func(remainingArgs []string, mc *MainConfig, sc *CastsConfig) error

var (
	CastsSections = map[string]CastsCliPoint{
		"help": RunCastHelp,
		"":     RunCastHelp,
		"list": func(remainingArgs []string, mc *MainConfig, sc *CastsConfig) error {
			profile, err := podcast_cdr_manager.OpenProfile(mc.profile)
			if err != nil {
				return fmt.Errorf("opening profile: %w", err)
			}
			casts, err := profile.ListCasts()
			if err != nil {
				return fmt.Errorf("getting casts: %w", err)
			}
			fmt.Printf("    %30s %30s %30s %s\n", "Publication Date", "On Disk", "Skipped?", "Title")
			for i, cast := range casts {
				fmt.Printf("% 3d %30s %30s %30s %s\n", i, cast.PubDate, cast.DiskName, fmt.Sprint(cast.SkippedDate), cast.Title)
			}
			return nil
		},
	}
)

func RunCastHelp(args []string, mc *MainConfig, sc *CastsConfig) error {
	fmt.Printf("usage: %s [options] [sections] [...]\n", os.Args[0])
	fmt.Printf("\toptions:\n")
	fmt.Printf("%19s %-20s %-39s\n", "-help", "", "This")
	fmt.Printf("\tsections:\n")
	fmt.Printf("%19s %-20s %-39s\n", "help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "list", "", "List casts")
	return nil
}

func RunCast(remainingArgs []string, mc *MainConfig) error {
	fs := flag.NewFlagSet("casts", flag.ExitOnError)
	help := fs.Bool("help", false, "")
	if err := fs.Parse(remainingArgs); err != nil {
		return fmt.Errorf("formatting args: %s", err)
	}
	sc := &CastsConfig{}
	if *help {
		if err := RunCastHelp(fs.Args(), mc, sc); err != nil {
			return fmt.Errorf("running help: %s", err)
		}
	}
	section, ok := CastsSections[fs.Arg(0)]
	if !ok {
		section = RunCastHelp
		fmt.Printf("Failed to find %s\n", fs.Arg(0))
	}
	if err := section(SkipFirstN(fs.Args(), 1), mc, sc); err != nil {
		return fmt.Errorf("running help: %s", err)
	}
	return nil
}
