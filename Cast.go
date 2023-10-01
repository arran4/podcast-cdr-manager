package podcast_cdr_manager

import "time"

type Cast struct {
	Title           string
	SubTitle        string
	Description     string
	SubscriptionUrl string
	Link            string
	MpegLink        string
	GUID            string
	Size            *int
	PubDate         *time.Time
	DiskName        string
	SkippedDate     *time.Time
}

func (p *Profile) ListCasts() ([]*Cast, error) {
	return p.Casts, nil
}
