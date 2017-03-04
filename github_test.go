package radar

import (
	"testing"
)

func TestJoinLinksIntoBody(t *testing.T) {
	header := "A new day! Here's what you have saved:\n\n"
	links := []RadarItem{{URL: "https://github.com"}}
	expected := header + "- [ ] [The world&#39;s leading software development platform · GitHub](https://github.com)\n"

	body, err := joinLinksIntoBody(links)
	if err != nil {
		t.Fatalf("Failed: expected err to be nil, but was %#v", err)
	}
	if body != expected {
		t.Fatalf("Failed: expected %#v, got %#v", expected, body)
	}
}
