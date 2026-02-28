package podcast_cdr_manager

import (
	"testing"
)

func TestProfileFindFreeDisksForSubscriptionBug(t *testing.T) {
	p := &Profile{
		Disks: []*Disk{
			{Name: "disk1", FilterPodcastUrl: []string{"http://test.com"}},
			{Name: "disk2", FilterPodcastUrl: []string{"http://other.com"}},
		},
	}
	free := p.FindFreeDisksForSubscription("http://test.com")

	if len(free) == 0 {
		t.Fatalf("Bug: Expected disk1, got 0 free disks. The filter deleted it!")
	}
}

func TestProfileFindFreeDisksForSubscriptionBug2(t *testing.T) {
	p := &Profile{
		Disks: []*Disk{
			{Name: "disk1", FilterPodcastUrl: []string{"http://test.com"}},
			{Name: "disk2", FilterPodcastUrl: []string{}},
		},
	}
	free := p.FindFreeDisksForSubscription("http://test.com")

	if len(free) != 2 {
		t.Fatalf("Bug: Expected disk1 and disk2 to be free for http://test.com, got %d free disks.", len(free))
	}
}
