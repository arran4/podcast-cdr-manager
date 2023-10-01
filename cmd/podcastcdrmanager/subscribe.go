package main

import (
	"flag"
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"os"
)

type SubscribeConfig struct{}

type SubscribeCliPoint func(remainingArgs []string, mc *MainConfig, sc *SubscribeConfig) error

var (
	SubscribeSections = map[string]SubscribeCliPoint{
		"help": RunSubscribeHelp,
		"":     RunSubscribeHelp,
		"rss": func(remainingArgs []string, mc *MainConfig, sc *SubscribeConfig) error {
			fs := flag.NewFlagSet("subscribe-rss", flag.ExitOnError)
			if err := fs.Parse(remainingArgs); err != nil {
				return fmt.Errorf("parsing arguments: %w", err)
			}
			profile, err := podcast_cdr_manager.OpenProfile(mc.profile)
			if err != nil {
				return fmt.Errorf("opening profile: %w", err)
			}
			n, err := profile.SubscribeToRss(fs.Arg(1))
			if err != nil {
				return fmt.Errorf("subscribing: %w", err)
			}
			if err := profile.Save(); err != nil {
				return fmt.Errorf("saving profile: %w", err)
			}
			fmt.Printf("Subscribed, added %d new items\n", n)
			return nil
		},
	}
)

func RunSubscribeHelp(args []string, mc *MainConfig, sc *SubscribeConfig) error {
	fmt.Printf("usage: %s [options] [sections] [...]\n", os.Args[0])
	fmt.Printf("\toptions:\n")
	fmt.Printf("%19s %-20s %-39s\n", "-help", "", "This")
	fmt.Printf("\tsections:\n")
	fmt.Printf("%19s %-20s %-39s\n", "help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "RunSubscribeRss", "<RSS URL>", "Subscribes to an RSS feed")
	return nil
}

func RunSubscribe(remainingArgs []string, mc *MainConfig) error {
	fs := flag.NewFlagSet("subscribe", flag.ExitOnError)
	help := fs.Bool("help", false, "")
	if err := fs.Parse(remainingArgs); err != nil {
		return fmt.Errorf("formatting args: %s", err)
	}
	sc := &SubscribeConfig{}
	if *help {
		if err := RunSubscribeHelp(fs.Args(), mc, sc); err != nil {
			return fmt.Errorf("running help: %s", err)
		}
	}
	section, ok := SubscribeSections[fs.Arg(1)]
	if !ok {
		section = RunSubscribeHelp
		fmt.Printf("Failed to find %s\n", fs.Arg(1))
	}
	if err := section(append([]string{fs.Arg(0)}, fs.Args()[2:]...), mc, sc); err != nil {
		return fmt.Errorf("running help: %s", err)
	}
	return nil
}
