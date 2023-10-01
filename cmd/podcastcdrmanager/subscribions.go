package main

import (
	"flag"
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"os"
)

type SubscribionsConfig struct{}

type SubscribionsCliPoint func(remainingArgs []string, mc *MainConfig, sc *SubscribionsConfig) error

var (
	SubscribionsSections = map[string]SubscribionsCliPoint{
		"help": RunSubscriptionHelp,
		"":     RunSubscriptionHelp,
		"list": func(remainingArgs []string, mc *MainConfig, sc *SubscribionsConfig) error {
			profile, err := podcast_cdr_manager.OpenProfile(mc.profile)
			if err != nil {
				return fmt.Errorf("opening profile: %w", err)
			}
			subs, err := profile.ListSubscriptions()
			if err != nil {
				return fmt.Errorf("getting subscriptions: %w", err)
			}
			for i, sub := range subs {
				fmt.Printf("%d %30s %s\n", i, sub.Name, sub.Url)
			}
			return nil
		},
	}
)

func RunSubscriptionHelp(args []string, mc *MainConfig, sc *SubscribionsConfig) error {
	fmt.Printf("usage: %s [options] [sections] [...]\n", os.Args[0])
	fmt.Printf("\toptions:\n")
	fmt.Printf("%19s %-20s %-39s\n", "-help", "", "This")
	fmt.Printf("\tsections:\n")
	fmt.Printf("%19s %-20s %-39s\n", "help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "RunSubscriptionRss", "<RSS URL>", "Subscribionss to an RSS feed")
	return nil
}

func RunSubscription(remainingArgs []string, mc *MainConfig) error {
	fs := flag.NewFlagSet("subscribions", flag.ExitOnError)
	help := fs.Bool("help", false, "")
	if err := fs.Parse(remainingArgs); err != nil {
		return fmt.Errorf("formatting args: %s", err)
	}
	sc := &SubscribionsConfig{}
	if *help {
		if err := RunSubscriptionHelp(fs.Args(), mc, sc); err != nil {
			return fmt.Errorf("running help: %s", err)
		}
	}
	section, ok := SubscribionsSections[fs.Arg(1)]
	if !ok {
		section = RunSubscriptionHelp
		fmt.Printf("Failed to find %s\n", fs.Arg(1))
	}
	if err := section(append([]string{fs.Arg(0)}, fs.Args()[2:]...), mc, sc); err != nil {
		return fmt.Errorf("running help: %s", err)
	}
	return nil
}
