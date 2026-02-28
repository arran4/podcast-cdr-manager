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
		if len(disk.FilterPodcastUrl) == 0 {
			return false
		}
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
	for j, word := range words {
		words[j] = strings.TrimSpace(word)
	}
	l := len(words) / 3
	if l == 0 {
		return fmt.Sprintf("disk-%d.iso", i)
	}

	ui := uint(i)
	if i < 0 {
		ui = uint(-i)
	}
	ul := uint(l)
	uw := uint(len(words))

	idx0 := int((ui%ul + (ul * ((ui / ul) % 3))) % uw)
	idx1 := int((ui%ul + (ul * ((ui/ul + 1) % 3))) % uw)
	idx2 := int((ui%ul + (ul * ((ui/ul + 2) % 3))) % uw)

	// Ensure we don't end up with identical words if len(words) is small
	if idx0 == idx1 { idx1 = (idx1 + 1) % len(words) }
	if idx0 == idx2 || idx1 == idx2 { idx2 = (idx2 + 1) % len(words) }
	if idx0 == idx2 || idx1 == idx2 { idx2 = (idx2 + 1) % len(words) }

	baseName := strings.Join([]string{
		words[idx0],
		words[idx1],
		words[idx2],
	}, "-")

	if ui >= ul {
		return fmt.Sprintf("%s-%d.iso", baseName, ui/ul)
	}
	return baseName + ".iso"
}

func createDiskIsoName(i int) string {
	words := strings.Split(wordsContent, "\n")
	for j, word := range words {
		words[j] = strings.TrimSpace(word)
	}
	if len(words) == 0 {
		return fmt.Sprintf("POD%d", i)
	}
	ui := uint(i)
	if i < 0 {
		ui = uint(-i)
	}
	return fmt.Sprintf("POD%s%d", strings.ToUpper(words[ui%uint(len(words))]), i)
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
