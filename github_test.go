package radar

import (
	"testing"
)

func TestJoinLinksIntoBody(t *testing.T) {
	header := "A new day! Here's what you have saved:"
	oldLinks := []RadarItem{{URL: "https://github.com"}}
	newLinks := []RadarItem{{URL: "https://byparker.com"}}
	expected := header + `

*Previously.* https://github.com/parkr/radar/issues/1

From before:

- [ ] [The world&#39;s leading software development platform · GitHub](https://github.com)


New:

- [ ] [Parker Moore - By Parker](https://byparker.com)
`

	body, err := generateBody(tmplData{
		OldIssueURL: "https://github.com/parkr/radar/issues/1",
		OldIssues:   oldLinks,
		NewIssues:   newLinks,
	})
	if err != nil {
		t.Fatalf("Failed: expected err to be nil, but was %#v", err)
	}
	if body != expected {
		t.Fatalf("Failed: expected %#v, got %#v", expected, body)
	}
}
