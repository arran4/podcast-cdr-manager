package podcast_cdr_manager

import (
	"testing"
)

func TestProfileGetSubByIndex(t *testing.T) {
	p := &Profile{
		Subscriptions: []*Subscription{
			{Url: "sub1"},
			{Url: "sub2"},
		},
	}
	sub, err := p.GetSubByIndex(0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Url != "sub1" {
		t.Fatalf("expected sub1, got %s", sub.Url)
	}

	_, err = p.GetSubByIndex(-1)
	if err == nil {
		t.Fatalf("expected error for negative index")
	}

	_, err = p.GetSubByIndex(2)
	if err == nil {
		t.Fatalf("expected error for out of bounds index")
	}
}
