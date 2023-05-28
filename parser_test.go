package radar

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractLinkedTodosFromMarkdown_oldStyleBody(t *testing.T) {
	body := `
A new day! Here's what you have saved:

[*Previously:*](https://github.com/parkr/daily/issues/1886)

- [ ] [Ankify: convert notes to Anki cards](https://ankify.krxiang.com/?utm_source=hackernewsletter&utm_medium=email&utm_term=show_hn)
- [ ] [WHERE TO WATCH BROADWAY ONLINE: THE THEATER LOVER’S GUIDE TO STREAMING](https://broadwaydirect.com/where-to-watch-musicals-online-the-musical-lovers-guide-to-streaming/)
- [ ] [Coding Relic](https://codingrelic.geekhold.com/?m=1)
- [ ] [“What's Next? (Black & White)” graphic tee, pullover hoodie, tank, and pullover crewneck by The West Wing Weekly. | Cotton Bureau](https://cottonbureau.com/products/whats-next-black-white#/4751515/tee-men-standard-tee-vintage-black-tri-blend-s)
- [ ] [“Parrot Style” graphic tee, pullover hoodie, onesie, tank, and pullover crewneck by deadpine. | Cotton Bureau](https://cottonbureau.com/products/parrot-style#/1941684/tee-men-standard-tee-white-100percent-cotton-s)
- [ ] [The West Wing Weekly | Cotton Bureau](https://cottonbureau.com/stores/the-west-wing-weekly#/shop)
- [ ] [Fools Crow - Wikipedia](https://en.m.wikipedia.org/wiki/Fools_Crow)
- [ ] [A Little Life - Wikipedia](https://en.m.wikipedia.org/wiki/A_Little_Life)
- [ ] [Archiving a website with wget](https://gist.github.com/mullnerz/9fff80593d6b442d5c1b)
- [ ] [gildas-lormeau/SingleFile: Web Extension for Firefox/Chrome/MS Edge and CLI tool to save a faithful copy of an entire web page in a single HTML file](https://github.com/gildas-lormeau/SingleFile?utm_source=hackernewsletter&utm_medium=email&utm_term=fav#projects-using-singlefile)
- [ ] [direnv/direnv: unclutter your .profile](https://github.com/direnv/direnv?utm_source=tldrnewsletter)
- [ ] [onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.](https://github.com/onceupon/Bash-Oneliner)
- [ ] [onceupon/Bash-Oneliner: A collection of handy Bash One-Liners and terminal tricks for data processing and Linux system maintenance.](https://github.com/onceupon/Bash-Oneliner)
- [ ] [dns.toys/main.go at master · knadh/dns.toys · GitHub](https://github.com/knadh/dns.toys/blob/master/cmd/dnstoys/main.go)
- [ ] [graham-essays: 📚 Download the full collection of Paul Graham essays in EPUB, Kindle & Markdown for easy reading.](https://github.com/ofou/graham-essays)
- [ ] [Humism Watches](https://humism.com/)
- [ ] [Song Notes](https://jamesfunk.net/songs-notes)
- [ ] [Read Hackernews on Kindle](https://ktool.io/hacker-news-to-kindle)
- [ ] [Email | Learn Netdata](https://learn.netdata.cloud/docs/agent/health/notifications/email)
- [ ] [Keynote: The Potential of Machine Learning for Hardware Design - Jeff Dean](https://m.youtube.com/watch?v=FraDFZ2t__A&t=269s&pp=2AGNApACAQ%3D%3D)
- [ ] [My simple note-taking setup | Zettelkasten in Obsidian | Step-by-step guide - YouTube](https://m.youtube.com/watch?v=E6ySG7xYgjY)
- [ ] [Why Retaining Walls Collapse - YouTube](https://m.youtube.com/watch?v=--DKkzWVh-E)
- [ ] [Heat Pumps are Not Hard: Here's what it will take to start pumping - YouTube](https://m.youtube.com/watch?v=43XKfuptnik)
- [ ] [Day-Date Day Wheels, Ethical Quandaries, Counterweights, And A Question You Shouldn't Ask - YouTube](https://m.youtube.com/watch?v=VAt8_ow91yI)
- [x] [The Mechanical Apple Watch | Watchfinder & Co. - YouTube](https://m.youtube.com/watch?v=BiPYOZnLJYo)
- [ ] [Sundar Pichai, CEO of Google and Alphabet - YouTube](https://m.youtube.com/watch?v=j9qGmO8Yy-Y)
- [ ] [Adam Falkner reads "Kissing Your Shoulder Blade Is the Most Honest Thing I've Done This Week" - YouTube](https://m.youtube.com/watch?v=QjkHEWFoEkY)
- [ ] [Little Shop of Horrors: Tiny Desk (Home) Concert - YouTube](https://m.youtube.com/watch?v=ymqKPz5kRXE)
- [ ] [Life is too short for dated CLI tools (Twitter thread)](https://mobile.twitter.com/amilajack/status/1479328649820000256)
- [ ] [Overcast auf Twitter: „Want to join the Overcast beta group? TestFlight: https://t.co/SQ97C8KmA0 Slack group for feedback, bug reports, and feature discussion: https://t.co/mC7rGQ43f1 Scammers sometimes charge for these links. Please don’t fall for it! Overcast’s beta is always free.“ / Twitter](https://mobile.twitter.com/OvercastFM/status/1514597131587313664)
- [ ] ["Shelter In Place" 5lb Bag : Ritual Coffee Roasters](https://ritualcoffee.com/shop/coffee/shelter-in-place-5lb/)


/cc @parkr
`
	items, err := extractLinkedTodosFromMarkdown(body)

	assert.NoError(t, err)
	assert.Equal(t, RadarItem{Title: "Ankify: convert notes to Anki cards", URL: "https://ankify.krxiang.com/?utm_source=hackernewsletter&utm_medium=email&utm_term=show_hn"}, items[0])
	assert.Equal(t, RadarItem{Title: `"Shelter In Place" 5lb Bag : Ritual Coffee Roasters`, URL: "https://ritualcoffee.com/shop/coffee/shelter-in-place-5lb/"}, items[len(items)-1])
	assert.Len(t, items, 30, "expecting 30 radar items")
}

func Test_extractLinkedTodosFromMarkdown_newLinkCommentBody(t *testing.T) {
	body := "- [ ] [changelog package - github.com/parkr/changelog@v1.1.0](https://pkg.go.dev/github.com/parkr/changelog@v1.1.0)"

	items, err := extractLinkedTodosFromMarkdown(body)

	assert.NoError(t, err)
	assert.Equal(t, []RadarItem{{Title: "changelog package - github.com/parkr/changelog@v1.1.0", URL: "https://pkg.go.dev/github.com/parkr/changelog@v1.1.0"}}, items)
}

func Test_extractLinkedTodosFromMarkdown_changelogFormat(t *testing.T) {
	body := `
A new day, @parkr! Here's what you have saved:

## New:

  * [ ] [Parker Moore | By Parker](https://byparker.com)
  * [ ] [Julia Evans](https://jvns.ca)

## *Previously:

  * [ ] [GitHub: Where the world builds software · GitHub](https://github.com)
  * [ ] [Ben Balter](https://ben.balter.com)

Previously: https://github.com/parkr/radar/issues/1
`
	expected := []RadarItem{
		{Title: "Parker Moore | By Parker", URL: "https://byparker.com"},
		{Title: "Julia Evans", URL: "https://jvns.ca"},
		{Title: "GitHub: Where the world builds software · GitHub", URL: "https://github.com"},
		{Title: "Ben Balter", URL: "https://ben.balter.com"},
	}

	items, err := extractLinkedTodosFromMarkdown(body)

	assert.NoError(t, err)
	assert.Equal(t, expected, items)
}

func Test_parseMarkdownLink(t *testing.T) {
	testcases := []struct {
		input string
		title string
		url   string
	}{
		{
			input: "[Ankify: convert notes to Anki cards](https://ankify.krxiang.com/?utm_source=hackernewsletter&utm_medium=email&utm_term=show_hn)",
			title: "Ankify: convert notes to Anki cards",
			url:   "https://ankify.krxiang.com/?utm_source=hackernewsletter&utm_medium=email&utm_term=show_hn",
		},
		{
			input: `[“What's Next? (Black & White)” graphic tee, pullover hoodie, tank, and pullover crewneck by The West Wing Weekly. | Cotton Bureau](https://cottonbureau.com/products/whats-next-black-white#/4751515/tee-men-standard-tee-vintage-black-tri-blend-s)`,
			title: `“What's Next? (Black & White)” graphic tee, pullover hoodie, tank, and pullover crewneck by The West Wing Weekly. | Cotton Bureau`,
			url:   "https://cottonbureau.com/products/whats-next-black-white#/4751515/tee-men-standard-tee-vintage-black-tri-blend-s",
		},
		{
			input: `[Little Shop of Horrors: Tiny Desk (Home) Concert - YouTube](https://m.youtube.com/watch?v=ymqKPz5kRXE)`,
			title: "Little Shop of Horrors: Tiny Desk (Home) Concert - YouTube",
			url:   "https://m.youtube.com/watch?v=ymqKPz5kRXE",
		},
		{
			input: `[Life is too short for dated CLI tools (Twitter thread)](https://mobile.twitter.com/amilajack/status/1479328649820000256)`,
			title: "Life is too short for dated CLI tools (Twitter thread)",
			url:   "https://mobile.twitter.com/amilajack/status/1479328649820000256",
		},
		{
			input: `[F) Sector 6 (NOC) - Metroid Fusion Guide - IGN](https://www.ign.com/wikis/metroid-fusion/F)_Sector_6_(NOC))`,
			title: "F) Sector 6 (NOC) - Metroid Fusion Guide - IGN",
			url:   "https://www.ign.com/wikis/metroid-fusion/F)_Sector_6_(NOC)",
		},
	}
	for _, testcase := range testcases {
		actualTitle, actualURL := parseMarkdownLink(testcase.input)
		assert.Equal(t, testcase.title, actualTitle, "Link: %q", testcase.input)
		assert.Equal(t, testcase.url, actualURL, "Link: %q", testcase.input)
	}
}

func Test_isBinaryResource(t *testing.T) {
	testcases := []struct {
		expected bool
		resp     *http.Response
		u        *url.URL
	}{
		{true, &http.Response{}, &url.URL{Path: "/hi.pdf"}},
		{false, &http.Response{}, &url.URL{Path: "/hi.html"}},
		{false, &http.Response{Header: http.Header{"Content-Type": []string{"text/html"}}}, &url.URL{}},
		{true, &http.Response{Header: http.Header{"Content-Type": []string{"text/plain"}}}, &url.URL{}},
	}
	for _, testcase := range testcases {
		actual := isBinaryResource(testcase.resp, testcase.u)
		if actual != testcase.expected {
			t.Fatalf("expected isBinaryResource to return %t, got %t\nresp: %#v\nurl: %#v", testcase.expected, actual, testcase.resp, testcase.u)
		}
	}
}

func Test_isPrivateHost(t *testing.T) {
	testcases := map[string]bool{
		"localhost":       true,
		"127.0.0.1":       true,  // IPv4 loopback
		"10.0.0.1":        true,  // RFC1918
		"172.16.0.1":      true,  // RFC1918
		"192.168.0.1":     true,  // RFC1918
		"169.254.0.1":     true,  // RFC3927 link-local
		"::1":             true,  // IPv6 loopback
		"fe80::":          true,  // IPv6 link-local
		"fc00::":          true,  // IPv6 unique local addr
		"129.129.129.129": false, // public
		"0.0.0.0":         false,
	}
	for input, expected := range testcases {
		actual := isPrivateHost(input)
		if actual != expected {
			t.Fatalf("expected %q to yield %t, got %t", input, expected, actual)
		}
	}
}

func Test_titleForWebpage(t *testing.T) {
	testcases := map[string]string{
		"http://localhost/foo/bar":        "A private page on localhost",
		"http://server.local/foo/bar":     "A private page on server.local",
		"https://myapp.localhost/foo/bar": "A private page on myapp.localhost",
		"http://127.0.0.1:1234/foo/bar":   "A private page on 127.0.0.1",
		"https://example.com":             "Example Domain",
	}
	for input, expected := range testcases {
		actual := titleForWebpage(input)
		if actual != expected {
			t.Fatalf("expected %q to yield %q, got %q", input, expected, actual)
		}
	}
}
