package main

import (
	"flag"
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
	"os"
)

type SubscriptionsConfig struct{}

type SubscriptionsCliPoint func(remainingArgs []string, mc *MainConfig, sc *SubscriptionsConfig) error

var (
	SubscriptionsSections = map[string]SubscriptionsCliPoint{
		"help": RunSubscriptionHelp,
		"":     RunSubscriptionHelp,
		"list": func(remainingArgs []string, mc *MainConfig, sc *SubscriptionsConfig) error {
			profile, err := podcast_cdr_manager.OpenProfile(mc.profile)
			if err != nil {
				return fmt.Errorf("opening profile: %w", err)
			}
			subs, err := profile.ListSubscriptions()
			if err != nil {
				return fmt.Errorf("getting subscriptions: %w", err)
			}
			for i, sub := range subs {
				fmt.Printf("% 3d %30s %s\n", i, sub.Name, sub.Url)
			}
			return nil
		},
		"refresh": func(remainingArgs []string, mc *MainConfig, sc *SubscriptionsConfig) error {
			profile, err := podcast_cdr_manager.OpenProfile(mc.profile)
			if err != nil {
				return fmt.Errorf("opening profile: %w", err)
			}
			subs, err := profile.ListSubscriptions()
			if err != nil {
				return fmt.Errorf("getting subscriptions: %w", err)
			}
			for i, sub := range subs {
				fmt.Printf("% 3d %30s %s\n", i, sub.Name, sub.Url)
				n, err := profile.UpdateSubscription(sub)
				if err != nil {
					return fmt.Errorf("updating subscription: %w", err)
				}
				fmt.Printf("%d New\n", n)
			}
			if err := profile.Save(); err != nil {
				return fmt.Errorf("saving profile: %w", err)
			}
			return nil
		},
	}
)

func RunSubscriptionHelp(args []string, mc *MainConfig, sc *SubscriptionsConfig) error {
	fmt.Printf("usage: %s [options] [sections] [...]\n", os.Args[0])
	fmt.Printf("\toptions:\n")
	fmt.Printf("%19s %-20s %-39s\n", "-help", "", "This")
	fmt.Printf("\tsections:\n")
	fmt.Printf("%19s %-20s %-39s\n", "help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "list", "", "List subscriptions")
	fmt.Printf("%19s %-20s %-39s\n", "refresh", "", "Refresh all subscriptions")
	return nil
}

func RunSubscription(remainingArgs []string, mc *MainConfig) error {
	fs := flag.NewFlagSet("subscriptions", flag.ExitOnError)
	help := fs.Bool("help", false, "")
	if err := fs.Parse(remainingArgs); err != nil {
		return fmt.Errorf("formatting args: %s", err)
	}
	sc := &SubscriptionsConfig{}
	if *help {
		if err := RunSubscriptionHelp(fs.Args(), mc, sc); err != nil {
			return fmt.Errorf("running help: %s", err)
		}
	}
	section, ok := SubscriptionsSections[fs.Arg(1)]
	if !ok {
		section = RunSubscriptionHelp
		fmt.Printf("Failed to find %s\n", fs.Arg(1))
	}
	if err := section(append([]string{fs.Arg(0)}, fs.Args()[min(2, len(fs.Args())):]...), mc, sc); err != nil {
		return fmt.Errorf("running help: %s", err)
	}
	return nil
}
