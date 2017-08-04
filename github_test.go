package radar

import (
	"testing"
)

func TestJoinLinksIntoBody(t *testing.T) {
	header := "A new day! Here's what you have saved:"
	oldLinks := []RadarItem{{URL: "https://github.com"}, {URL: "https://ben.balter.com", Title: "Ben Balter"}}
	newLinks := []RadarItem{{URL: "https://byparker.com"}, {URL: "https://jvns.ca"}}
	expected := header + `

[*Previously:*](https://github.com/parkr/radar/issues/1)

- [ ] [The world&#39;s leading software development platform · GitHub](https://github.com)
- [ ] [Ben Balter](https://ben.balter.com)

New:

- [ ] [Parker Moore - By Parker](https://byparker.com)
- [ ] [Julia Evans - Julia Evans](https://jvns.ca)

/cc @parkr
`

	body, err := generateBody(&tmplData{
		OldIssueURL: "https://github.com/parkr/radar/issues/1",
		OldIssues:   oldLinks,
		NewIssues:   newLinks,
		Mention:     "@parkr",
	})
	if err != nil {
		t.Fatalf("Failed: expected err to be nil, but was %#v", err)
	}
	if body != expected {
		t.Fatalf("Failed: expected %#v, got %#v", expected, body)
	}
}
