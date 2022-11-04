package radar

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinLinksIntoBody(t *testing.T) {
	header := "A new day, @parkr! Here's what you have saved:"
	oldLinks := []RadarItem{{URL: "https://github.com"}, {URL: "https://ben.balter.com", Title: "Ben Balter"}}
	newLinks := []RadarItem{{URL: "https://byparker.com"}, {URL: "https://jvns.ca"}}
	expected := header + `

## New:

  * [ ] [Parker Moore | By Parker](https://byparker.com)
  * [ ] [Julia Evans](https://jvns.ca)

## [*Previously:*](https://github.com/parkr/radar/issues/1)

  * [ ] [GitHub: Let’s build from here · GitHub](https://github.com)
  * [ ] [Ben Balter](https://ben.balter.com)
`

	body, err := generateBody(&tmplData{
		OldIssueURL: "https://github.com/parkr/radar/issues/1",
		OldLinks:    oldLinks,
		NewLinks:    newLinks,
		Mention:     "@parkr",
	})
	if err != nil {
		t.Fatalf("Failed: expected err to be nil, but was %#v", err)
	}
	if body != expected {
		assert.Equal(t, expected, body)
		t.Fatalf("Failed: expected\n\n%s\n\n, got:\n\n%s", expected, body)
	}
}
