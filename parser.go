package radar

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"regexp"
)

var titleExtractorRegexp = regexp.MustCompile("(?i)<title>(.+)</title>")
var markdownLinkExtractorRegexp = regexp.MustCompile("-\\s+\\[ \\]\\s+\\[(.+)\\]\\((.+)\\)")

func (r RadarItem) GetTitle() string {
	if r.Title != "" {
		return r.Title
	}

	u, err := url.Parse(r.URL)
	if err != nil {
		return r.URL
	}

	if isBinaryResource(u) {
		return "File on " + u.Hostname()
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return "A page on " + u.Hostname()
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	matches := titleExtractorRegexp.FindAllStringSubmatch(string(body), -1)
	if len(matches) < 1 || len(matches[0]) < 2 {
		return "A page on " + u.Hostname()
	}
	return matches[0][1]
}

var parsableExtensions = map[string]bool{
	"":       true,
	".html":  true,
	".xhtml": true,
	".htm":   true,
}

func isBinaryResource(u *url.URL) bool {
	ext := path.Ext(u.Path)
	_, ok := parsableExtensions[ext]
	return !ok
}

func extractLinkedTodosFromMarkdown(body string) []RadarItem {
	var items []RadarItem
	for _, match := range markdownLinkExtractorRegexp.FindAllStringSubmatch(body, -1) {
		if len(match) < 3 {
			continue
		}
		items = append(items, RadarItem{Title: match[1], URL: match[2]})
	}
	return items
}
