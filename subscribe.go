package podcast_cdr_manager

import (
	"fmt"
	"github.com/mmcdole/gofeed"
	"net/http"
)

func (p *Profile) SubscribeToRss(url string) (int, error) {
	if p.HasSubscriptionByUrl(url) {
		return 0, fmt.Errorf("already have subscription")
	}
	feed, err := p.GetFeed(url)
	if err != nil {
		return 0, err
	}
	sub := &Subscription{
		Name: feed.Title,
		Url:  url,
		Type: "rss",
	}
	p.Subscriptions = append(p.Subscriptions, sub)
	return p.UpdateSubscriptionWithFeed(sub, feed)
}

func (p *Profile) GetFeed(url string) (*gofeed.Feed, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	switch r.StatusCode {
	case http.StatusOK:
	default:
		return nil, fmt.Errorf("got http returned: %d: %s", r.StatusCode, r.Status)
	}
	fp := gofeed.NewParser()
	feed, err := fp.Parse(r.Body)
	if err != nil {
		return nil, fmt.Errorf("http read / feed parse: %w", err)
	}
	return feed, nil
}
