package podcast_cdr_manager

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"
	"time"
)

type Disk struct {
	Filename         string
	Name             string
	CreatedDate      *time.Time
	BurntDate        *time.Time
	UsedSpaceMb      int
	TotalSpaceMb     int
	FilterPodcastUrl []string
	ReadyToBurn      *time.Time
}

func (p *Profile) ListDisks() ([]*Disk, error) {
	return p.Disks, nil
}

func (p *Profile) FindFreeDisksForSubscription(subscriptionUrl string) []*Disk {
	return slices.DeleteFunc(p.FindFreeDisks(), func(disk *Disk) bool {
		for _, u := range disk.FilterPodcastUrl {
			if strings.EqualFold(u, subscriptionUrl) {
				return false
			}
		}
		return true
	})
}

func (p *Profile) FindFreeDiskForSubscription(subscriptionUrl string) (*Disk, error) {
	disks := p.FindFreeDisksForSubscription(subscriptionUrl)
	if len(disks) == 0 {
		return nil, nil
	}
	return disks[0], nil
}
func (p *Profile) FindFreeDisk() (*Disk, error) {
	disks := p.FindFreeDisks()
	if len(disks) == 0 {
		return nil, nil
	}
	return disks[0], nil
}

func (p *Profile) GetDiskByIndex(index int) (*Disk, error) {
	if 0 > index {
		return nil, fmt.Errorf("please select a disk")
	}
	if len(p.Disks) <= index {
		return nil, fmt.Errorf("no such disk")
	}
	return p.Disks[index], nil
}

func (p *Profile) FindFreeDisks() []*Disk {
	return slices.DeleteFunc(slices.Clone(p.Disks), func(disk *Disk) bool {
		return disk.BurntDate != nil || disk.ReadyToBurn != nil
	})
}

var (
	//go:embed "words.txt"
	wordsContent string
)

func createDiskFilename(i int) string {
	words := strings.Split(wordsContent, "\n")
	for i, word := range words {
		words[i] = strings.TrimSpace(word)
	}
	// TODO handle overflow intelligently
	l := len(words) / 3
	return strings.Join([]string{
		words[i%l+(l*(i/l+0))],
		words[i%l+(l*(i/l+1))],
		words[i%l+(l*(i/l+2))],
	}, "-") + ".iso"
}

func createDiskIsoName(i int) string {
	words := strings.Split(wordsContent, "\n")
	for i, word := range words {
		words[i] = strings.TrimSpace(word)
	}
	return fmt.Sprintf("POD%s%d", strings.ToUpper(words[i]), i)
}

func (p *Profile) CreateDisk(subscriptionUrlFilter []string, diskSizeMb int) (*Disk, error) {
	now := time.Now()
	d := &Disk{
		Name:             createDiskIsoName(len(p.Disks)),
		Filename:         createDiskFilename(len(p.Disks)),
		CreatedDate:      &now,
		BurntDate:        nil,
		UsedSpaceMb:      1, // We will use some space.
		TotalSpaceMb:     diskSizeMb,
		FilterPodcastUrl: subscriptionUrlFilter,
		ReadyToBurn:      nil,
	}
	p.Disks = append(p.Disks, d)
	return d, nil
}
