package commands

import (
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
)

// Subscribe is a subcommand `podcast-cdr-manager subscribe`
// parent-flag: profile
func Subscribe(profile string) {
}

// SubscribeRss is a subcommand `podcast-cdr-manager subscribe rss`
// parent-flag: profile
// Flags:
//   url: @1 The RSS URL
func SubscribeRss(profile string, url string) error {
	profile = GetProfile(profile)
	p, err := podcast_cdr_manager.OpenProfile(profile)
	if err != nil {
		return fmt.Errorf("opening profile: %w", err)
	}
	n, err := p.SubscribeToRss(url)
	if err != nil {
		return fmt.Errorf("subscribing: %w", err)
	}
	if err := p.Save(); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}
	fmt.Printf("Subscribed, added %d new items\n", n)
	return nil
}
