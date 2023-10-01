package podcast_cdr_manager

import "time"

type Cast struct {
	Title            string
	SubTitle         string
	Description      string
	SubscriptionUrl  string
	Link             string
	MpegLink         string
	GUID             string
	Size             *int
	ISOName          string
	PubDate          *time.Time
	PlannedDate      *time.Time
	ISOGeneratedDate []time.Time
	SkippedDate      *time.Time
}
