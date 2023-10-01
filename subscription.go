package podcast_cdr_manager

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"sort"
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

func (p *Profile) GetSubByIndex(index int) (*Subscription, error) {
	if len(p.Subscriptions) <= index {
		return nil, fmt.Errorf("no such subscription")
	}
	return p.Subscriptions[index], nil
}

func (p *Profile) HasSubscriptionByUrl(url string) bool {
	for _, s := range p.Subscriptions {
		if strings.EqualFold(s.Url, url) {
			return true
		}
	}
	return false
}

func (p *Profile) RefreshSubscription(sub *Subscription) (int, error) {
	feed, err := p.GetFeed(sub.Url)
	if err != nil {
		return 0, err
	}
	return p.RefreshSubscriptionWithFeed(sub, feed)
}

type ByPubDate []*gofeed.Item

func (b ByPubDate) Len() int {
	return len(b)
}

func (b ByPubDate) Less(i, j int) bool {
	if b[i].PublishedParsed == nil {
		return true
	}
	if b[j].PublishedParsed == nil {
		return true
	}
	return b[i].PublishedParsed.Before(*b[j].PublishedParsed)
}

func (b ByPubDate) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (p *Profile) RefreshSubscriptionWithFeed(sub *Subscription, feed *gofeed.Feed) (int, error) {
	existingItemMap := map[string]*Cast{}
	for _, cast := range p.Casts {
		existingItemMap[cast.GUID] = cast
	}
	added := 0
	items := feed.Items
	sort.Sort(ByPubDate(items))
	for _, item := range items {
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
			Title:           item.Title,
			SubTitle:        item.Custom["itunes:subtitle"],
			Description:     item.Description,
			SubscriptionUrl: sub.Url,
			Link:            item.Link,
			MpegLink:        mpegLink,
			GUID:            item.GUID,
			SizeBytes:       size,
			PubDate:         item.PublishedParsed,
			SkippedDate:     nil,
		})
	}
	return added, nil
}
