package podcast_cdr_manager

import (
	"slices"
	"time"
)

type Cast struct {
	Title           string
	SubTitle        string
	Description     string
	SubscriptionUrl string
	Link            string
	MpegLink        string
	GUID            string
	SizeBytes       *int
	PubDate         *time.Time
	DiskName        string
	SkippedDate     *time.Time
}

func (p *Profile) ListCasts() ([]*Cast, error) {
	return p.Casts, nil
}

func (p *Profile) ListUnassignedCasts() ([]*Cast, error) {
	return slices.DeleteFunc(slices.Clone(p.Casts), func(cast *Cast) bool {
		return cast.SkippedDate != nil || cast.DiskName != ""
	}), nil
}
