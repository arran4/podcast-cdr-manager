package podcast_cdr_manager

import (
	"testing"
	"time"
)

func TestCreateDiskFilename(t *testing.T) {
	// Let's call createDiskFilename with increasing values of i
	// to see if we can trigger the panic.
	for i := 0; i < 10000; i++ {
		_ = createDiskFilename(i)
	}
}

func TestCreateDiskIsoName(t *testing.T) {
	for i := 0; i < 10000; i++ {
		_ = createDiskIsoName(i)
	}
}

func TestFindFreeDisks(t *testing.T) {
	p := &Profile{
		Disks: []*Disk{
			{Name: "disk1"}, // free
			{Name: "disk2", BurntDate: new(time.Time)}, // burnt
			{Name: "disk3", ReadyToBurn: new(time.Time)}, // ready to burn
		},
	}
	free := p.FindFreeDisks()
	if len(free) != 1 {
		t.Fatalf("expected 1 free disk, got %d", len(free))
	}
	if free[0].Name != "disk1" {
		t.Fatalf("expected disk1, got %s", free[0].Name)
	}
}

func TestFindFreeDisksForSubscription(t *testing.T) {
	p := &Profile{
		Disks: []*Disk{
			{Name: "disk1", FilterPodcastUrl: []string{"http://test.com"}},
			{Name: "disk2", FilterPodcastUrl: []string{"http://other.com"}},
		},
	}
	free := p.FindFreeDisksForSubscription("http://test.com")
	if len(free) != 1 {
		t.Fatalf("expected 1 free disk, got %d", len(free))
	}
	if free[0].Name != "disk1" {
		t.Fatalf("expected disk1, got %s", free[0].Name)
	}
}
