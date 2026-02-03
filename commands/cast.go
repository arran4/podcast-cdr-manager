package commands

import (
	"fmt"
	podcast_cdr_manager "github.com/arran4/podcast-cdr-manager"
)

// Cast is a subcommand `podcast-cdr-manager cast`
// parent-flag: profile
func Cast(profile string) {
}

// CastList is a subcommand `podcast-cdr-manager cast list-casts`
// parent-flag: profile
func CastList(profile string) error {
	profile = GetProfile(profile)
	p, err := podcast_cdr_manager.OpenProfile(profile)
	if err != nil {
		return fmt.Errorf("opening profile: %w", err)
	}
	casts, err := p.ListCasts()
	if err != nil {
		return fmt.Errorf("getting casts: %w", err)
	}
	fmt.Printf("    %30s %30s %30s %s\n", "Publication Date", "On Disk", "Skipped?", "Title")
	for i, cast := range casts {
		fmt.Printf("% 3d %30s %30s %30s %s\n", i, cast.PubDate, cast.DiskName, fmt.Sprint(cast.SkippedDate), cast.Title)
	}
	return nil
}
