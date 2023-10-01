package podcast_cdr_manager

import (
	"github.com/mmcdole/gofeed"
	"strconv"
	"strings"
)

type Subscription struct {
	Name string
	Url  string
	Type string
}

func (p *Profile) ListSubscriptions() ([]*Subscription, error) {
	return p.Subscriptions, nil
}

func (p *Profile) HasSubscriptionByUrl(url string) bool {
	for _, s := range p.Subscriptions {
		if strings.EqualFold(s.Url, url) {
			return true
		}
	}
	return false
}

func (p *Profile) UpdateSubscription(sub *Subscription) (int, error) {
	feed, err := p.GetFeed(sub.Url)
	if err != nil {
		return 0, err
	}
	return p.UpdateSubscriptionWithFeed(sub, feed)
}

func (p *Profile) UpdateSubscriptionWithFeed(sub *Subscription, feed *gofeed.Feed) (int, error) {
	existingItemMap := map[string]*Cast{}
	for _, cast := range p.Casts {
		existingItemMap[cast.GUID] = cast
	}
	added := 0
	for _, item := range feed.Items {
		if _, found := existingItemMap[item.GUID]; found {
			continue
		}
		var mpegLink string
		var size *int
		for _, enc := range item.Enclosures {
			switch enc.Type {
			case "audio/mpeg":
				mpegLink = enc.URL
				if v, err := strconv.ParseInt(enc.Length, 10, 64); err == nil {
					v := int(v)
					size = &v
				}
			}
		}
		added++
		p.Casts = append(p.Casts, &Cast{
			Title:            item.Title,
			SubTitle:         item.Custom["itunes:subtitle"],
			Description:      item.Description,
			SubscriptionUrl:  sub.Url,
			Link:             item.Link,
			MpegLink:         mpegLink,
			GUID:             item.GUID,
			Size:             size,
			ISOName:          "",
			PubDate:          item.PublishedParsed,
			PlannedDate:      nil,
			ISOGeneratedDate: nil,
			SkippedDate:      nil,
		})
	}
	return added, nil
}
