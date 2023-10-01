package podcast_cdr_manager

import (
	"time"
)

type Disk struct {
	Filename    string
	CreatedDate *time.Time
	BurntDate   *time.Time
}

func (p *Profile) ListDisks() ([]*Disk, error) {
	return p.Disks, nil
}
