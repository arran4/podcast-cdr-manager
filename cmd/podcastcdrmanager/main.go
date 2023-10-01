package main

import (
	"flag"
	"fmt"
	"os"
)

type MainConfig struct {
	profile string
}

type MainCliPoint func(remainingArgs []string, mc *MainConfig) error

var (
	MainSections = map[string]MainCliPoint{
		"subscribe":    RunSubscribe,
		"subscription": RunSubscription,
		"profile":      RunProfile,
		"help":         RunHelp,
		"":             RunHelp,
	}
)

func RunHelp(args []string, mc *MainConfig) error {
	fmt.Printf("usage: %s [options] [sections] [...]\n", os.Args[0])
	fmt.Printf("\toptions:\n")
	fmt.Printf("%19s %-20s %-39s\n", "-help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "-profile", "[string:default]", "The user profile to use")
	fmt.Printf("\tsections:\n")
	fmt.Printf("%19s %-20s %-39s\n", "help", "", "This")
	fmt.Printf("%19s %-20s %-39s\n", "subscribe", "", "Subscribe to a new podcast")
	fmt.Printf("%19s %-20s %-39s\n", "subscriptions", "", "Manage existing subscriptions")
	fmt.Printf("%19s %-20s %-39s\n", "profile", "", "Profile management")
	return nil
}

func main() {
	fs := flag.NewFlagSet("profile", flag.ExitOnError)
	profile := fs.String("profile", "default", "The profile/user")
	help := fs.Bool("help", false, "")
	if err := fs.Parse(os.Args); err != nil {
		fmt.Printf("Error formatting args: %s\n", err)
		os.Exit(-1)
		return
	}
	mc := &MainConfig{
		profile: *profile,
	}
	if *help {
		if err := RunHelp(fs.Args(), mc); err != nil {
			fmt.Printf("Error running help: %s\n", err)
			os.Exit(-1)
			return
		}
		return
	}
	section, ok := MainSections[fs.Arg(1)]
	if !ok {
		section = RunHelp
		fmt.Printf("Failed to find %s\n", fs.Arg(1))
	}
	if err := section(append([]string{fs.Arg(0)}, fs.Args()[2:]...), mc); err != nil {
		fmt.Printf("Error running %s: %s\n", flag.Arg(1), err)
		os.Exit(-1)
		return
	}

}
