package radar

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var titleExtractorRegexp = regexp.MustCompile("(?i)<title>(.+)</title>")
var markdownLinkExtractorRegexp = regexp.MustCompile("-\\s+\\[ \\]\\s+\\[(.+)\\]\\((.+)\\)")

func (r RadarItem) GetTitle() string {
	if r.Title == "" {
		r.Title = titleForWebpage(r.URL)
	}
	return r.Title
}

func titleForWebpage(urlString string) string {
	u, err := url.Parse(urlString)
	if err != nil {
		return urlString
	}

	if u.Hostname() == "github.com" && u.Path != "" {
		if title := titleForGitHubReference(u); title != "" {
			return title
		}
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

func titleForGitHubReference(u *url.URL) string {
	// Oof.
	client := getClient(os.Getenv("GITHUB_ACCESS_TOKEN"))
	ctx := context.Background()

	// /parkr/radar/issues/1
	// /parkr/radar/issues
	// /parkr/radar/pulls
	// /parkr/radar/pull/2
	pieces := strings.Split(u.Path, "/")
	switch len(pieces) {
	case 2:
		// Repo, e.g. /parkr/radar
		owner, name := pieces[0], pieces[1]
		if repo, _, err := client.Repositories.Get(ctx, owner, name); err == nil {
			return fmt.Sprintf("%s/%s: %s", owner, name, *repo.Description)
		}
	case 4:
		// Subpage of a repo, e.g. /parkr/radar/blob/master, */issues/1, */pull/2, etc
		number, err := strconv.Atoi(pieces[3])
		if err != nil {
			return ""
		}

		owner, name := pieces[0], pieces[1]
		if pieces[2] == "issues" {
			// Get issue title from API
			if issue, _, err := client.Issues.Get(ctx, owner, name, number); err == nil {
				return fmt.Sprintf("%s - Issue #%d - %s/%s", *issue.Title, number, owner, name)
			}
		}
		if pieces[2] == "pull" {
			// Get pull title from API
			if pr, _, err := client.PullRequests.Get(ctx, owner, name, number); err == nil {
				return fmt.Sprintf("%s - Pull request #%d - %s/%s", *pr.Title, number, owner, name)
			}
		}
	}

	return ""
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
