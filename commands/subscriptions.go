package commands

import (
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
)

// Subscriptions is a subcommand `podcast-cdr-manager subscriptions`
// parent-flag: profile
func Subscriptions(profile string) {
}

// SubscriptionsList is a subcommand `podcast-cdr-manager subscriptions list-subs`
// parent-flag: profile
func SubscriptionsList(profile string) error {
	profile = GetProfile(profile)
	p, err := podcast_cdr_manager.OpenProfile(profile)
	if err != nil {
		return fmt.Errorf("opening profile: %w", err)
	}
	subs, err := p.ListSubscriptions()
	if err != nil {
		return fmt.Errorf("getting subscriptions: %w", err)
	}
	for i, sub := range subs {
		fmt.Printf("% 3d %30s %s\n", i, sub.Name, sub.Url)
	}
	return nil
}

// SubscriptionsRefresh is a subcommand `podcast-cdr-manager subscriptions refresh`
// parent-flag: profile
func SubscriptionsRefresh(profile string) error {
	profile = GetProfile(profile)
	p, err := podcast_cdr_manager.OpenProfile(profile)
	if err != nil {
		return fmt.Errorf("opening profile: %w", err)
	}
	subs, err := p.ListSubscriptions()
	if err != nil {
		return fmt.Errorf("getting subscriptions: %w", err)
	}
	for i, sub := range subs {
		fmt.Printf("% 3d %30s %s\n", i, sub.Name, sub.Url)
		n, err := p.RefreshSubscription(sub)
		if err != nil {
			return fmt.Errorf("updating subscription: %w", err)
		}
		fmt.Printf("%d New\n", n)
	}
	if err := p.Save(); err != nil {
		return fmt.Errorf("saving profile: %w", err)
	}
	return nil
}
