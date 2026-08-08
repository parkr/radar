package radar

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/parkr/changelog"
)

var titleExtractorRegexp = regexp.MustCompile("(?i)<title>(.+)</title>")

func (r RadarItem) GetTitle() string {
	if r.Title == "" {
		r.Title = titleForWebpage(r.URL)
	}
	return r.Title
}

func titleForWebpage(urlString string) string {
	inputURL, err := url.Parse(urlString)
	if err != nil {
		return urlString
	}

	if inputURL.Hostname() == "github.com" && inputURL.Path != "" {
		if title := titleForGitHubReference(inputURL); title != "" {
			return title
		}
	}

	if isPrivateHost(inputURL.Hostname()) {
		return "A private page on " + inputURL.Hostname()
	}

	u := &url.URL{
		Scheme:   inputURL.Scheme,
		Host:     inputURL.Host,
		Path:     inputURL.Path,
		RawQuery: inputURL.RawQuery,
		Fragment: inputURL.Fragment,
	}

	resp, err := http.Get(u.String())
	if err != nil {
		return "A page on " + u.Hostname()
	}
	if isBinaryResource(resp, u) {
		return "File on " + u.Hostname()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "A page on " + u.Hostname()
	}
	matches := titleExtractorRegexp.FindAllStringSubmatch(string(body), -1)
	if len(matches) < 1 || len(matches[0]) < 2 {
		return "A page on " + u.Hostname()
	}
	return matches[0][1]
}

func titleForGitHubReference(u *url.URL) string {
	// Oof.
	client := NewGitHubClient(os.Getenv("GITHUB_ACCESS_TOKEN"))
	ctx := context.Background()

	// Trim /files from the end and strip / from the beginning.
	path := strings.TrimPrefix(strings.TrimSuffix(u.Path, "/files"), "/")

	// /parkr/radar/issues/1
	// /parkr/radar/issues
	// /parkr/radar/pulls
	// /parkr/radar/pull/2
	pieces := strings.Split(path, "/")
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

var parsableContentTypes = []string{
	"text/html", "text/xhtml", "text/xml",
}

func isBinaryResource(resp *http.Response, u *url.URL) bool {
	if contentType := resp.Header.Get("Content-Type"); contentType != "" {
		for _, parsableContentType := range parsableContentTypes {
			if strings.Contains(contentType, parsableContentType) {
				return false
			}
		}
		return true
	}

	ext := path.Ext(u.Path)
	_, ok := parsableExtensions[ext]
	return !ok
}

func extractLinkedTodosFromMarkdown(body string) ([]RadarItem, error) {
	var items []RadarItem
	chlog, err := changelog.NewChangelogFromReader(strings.NewReader(body))
	if err != nil {
		return items, err
	}
	for _, version := range chlog.Versions {
		for _, line := range version.History {
			// Checked off, ignore.
			if strings.HasPrefix(line.Summary, "[x]") || strings.HasPrefix(line.Summary, "[X]") {
				continue
			}
			// Not checked off, parse and include.
			linkText := line.Summary[len("[ ] "):]
			title, url, rawSuffix := parseMarkdownLinkFull(linkText)
			if url != "" {
				items = append(items, RadarItem{Title: title, URL: url, RawSuffix: trimSuffix(rawSuffix)})
			} else {
				Printf("unable to parse link [skip]: %s", linkText)
			}
		}
	}
	return items, nil
}

// parseMarkdownLink parses a markdown link of the form [title](url) and returns
// the title and url. Any trailing text after the closing ) is ignored.
func parseMarkdownLink(link string) (title string, url string) {
	title, url, _ = parseMarkdownLinkFull(link)
	return
}

// parseMarkdownLinkFull parses a markdown link of the form [title](url) and
// returns the title, url, and any trailing text after the closing ).
// The suffix may be any text: plain notes, italic annotations, or multi-line
// continuations added by the user.
func parseMarkdownLinkFull(link string) (title, url, suffix string) {
	boundaryIdx := strings.LastIndex(link, "](")
	openingIdx := strings.Index(link, "[")
	if boundaryIdx < 0 || openingIdx < 0 {
		return "", "", ""
	}
	closingParenIdx := findURLClosingParen(link, boundaryIdx+2)
	if closingParenIdx < 0 {
		return "", "", ""
	}
	title = link[openingIdx+1 : boundaryIdx]
	url = link[boundaryIdx+2 : closingParenIdx]
	if closingParenIdx+1 < len(link) {
		suffix = link[closingParenIdx+1:]
	}
	return
}

// findURLClosingParen finds the index of the ) that closes the markdown link URL
// (i.e. the portion that started with ]( at urlStart-2). It uses a heuristic:
// the first ) that is followed by end-of-string, whitespace, or another
// suffix-starting character. If no such ) is found it falls back to the last )
// in the string (for URLs that themselves contain unescaped ) characters and
// have no trailing suffix).
func findURLClosingParen(link string, urlStart int) int {
	for i := urlStart; i < len(link); i++ {
		if link[i] != ')' {
			continue
		}
		// ) at end of string, or followed by whitespace → this closes the URL.
		if i == len(link)-1 || link[i+1] == ' ' || link[i+1] == '\t' || link[i+1] == '\n' || link[i+1] == '\r' {
			return i
		}
	}
	// Fallback: the last ) in the string closes the URL (handles URLs that
	// themselves end with ) and have no suffix).
	return strings.LastIndex(link, ")")
}

// trimSuffix removes any structurally-spurious tail from a raw suffix. The
// changelog library may attach non-item lines (section headers, footers) to the
// last bullet in a section. We keep the same-line portion and any indented
// continuation lines, but stop at the first unindented non-empty line, which
// signals an issue-format element rather than a user annotation.
func trimSuffix(suffix string) string {
	if !strings.Contains(suffix, "\n") {
		return suffix
	}
	lines := strings.Split(suffix, "\n")
	result := lines[0]
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		switch {
		case line == "":
			result += "\n" + line
		case len(line) > 0 && (line[0] == ' ' || line[0] == '\t'):
			result += "\n" + line
		default:
			// Unindented non-empty line: issue structural content, not annotation.
			goto done
		}
	}
done:
	return strings.TrimRight(result, "\n")
}

var privateIPBlocks []*net.IPNet

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func isPrivateHost(hostname string) bool {
	addr := net.ParseIP(hostname)
	if addr != nil {
		return isPrivateIP(addr)
	}

	return hostname == "localhost" || strings.HasSuffix(hostname, ".local") || strings.HasSuffix(hostname, ".localhost")
}
