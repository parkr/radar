package radar

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/parkr/changelog"
	"golang.org/x/oauth2"
)

// Generate and re-use one client per token. Key = token, value = client for token.
var clients = map[string]*github.Client{}

var labels = []string{"radar"}

type tmplData struct {
	OldIssueURL string
	NewLinks    []RadarItem
	OldLinks    []RadarItem
	Mention     string
}

func GenerateRadarIssue(radarItemsService RadarItemsService, mention string) (*github.Issue, error) {
	client := radarItemsService.githubClient
	owner, name := radarItemsService.owner, radarItemsService.repoName
	var err error

	data := &tmplData{
		Mention: mention,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	previousIssue := getPreviousRadarIssue(ctx, client, owner, name)
	if previousIssue != nil {
		data.OldIssueURL = *previousIssue.HTMLURL
		data.OldLinks, data.NewLinks, err = extractGitHubLinks(ctx, client, owner, name, previousIssue)
		if err != nil {
			Printf("Unable to extract GitHub links from %s/%s#%d", owner, name, *previousIssue.Number)
			return nil, err
		}
	}

	sort.Stable(RadarItems(data.NewLinks))
	sort.Stable(RadarItems(data.OldLinks))

	body, err := generateBody(data)
	if err != nil {
		Printf("Couldn't get a radar body: %#v", err)
		return nil, err
	}

	newIssue, _, err := client.Issues.Create(ctx, owner, name, &github.IssueRequest{
		Title:  github.String(getTitle()),
		Body:   github.String(body),
		Labels: &labels,
	})
	if err != nil {
		return nil, err
	}

	// Close old issue.
	if previousIssue != nil {
		_, _, err := client.Issues.Edit(
			ctx, owner, name, *previousIssue.Number, &github.IssueRequest{State: github.String("closed")},
		)
		if err != nil {
			Printf("%s/%s: error closing issue number=%d: %#v", owner, name, *previousIssue.Number, err)
		}
	}

	return newIssue, nil
}

func getPreviousRadarIssue(ctx context.Context, client *github.Client, owner, name string) *github.Issue {
	query := fmt.Sprintf("repo:%s/%s is:open is:issue label:radar", owner, name)
	opts := &github.SearchOptions{
		Sort:        "created",
		Order:       "desc",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	result, _, err := client.Search.Issues(ctx, query, opts)
	if err != nil {
		Printf("Error running query '%s': %#v", query, err)
		return nil
	}

	if len(result.Issues) == 0 {
		Printf("No issues for '%s'.", query)
		return nil
	}

	return result.Issues[0]
}

func getTitle() string {
	return fmt.Sprintf("Radar for %s", time.Now().Format("2006-01-02"))
}

func generateBody(data *tmplData) (string, error) {
	if len(data.NewLinks) == 0 && len(data.OldLinks) == 0 {
		return "Nothing to do today. Nice work! :sparkles:", nil
	}

	buf := bytes.NewBufferString("A new day, " + data.Mention + "! Here's what you have saved:\n\n")
	links := changelog.NewChangelog()
	for _, newIssue := range data.NewLinks {
		links.AddLineToVersion("New:", &changelog.ChangeLine{Summary: "[ ] " + newIssue.GetMarkdown()})
	}
	previouslyHeader := "*Previously:*"
	for _, oldIssue := range data.OldLinks {
		links.AddLineToVersion(previouslyHeader, &changelog.ChangeLine{Summary: "[ ] " + oldIssue.GetMarkdown()})
	}
	fmt.Fprintf(buf, links.String())
	if data.OldIssueURL != "" {
		fmt.Fprintf(buf, "\n*Previously:* %s\n", data.OldIssueURL)
	}
	return buf.String(), nil
}

func extractGitHubLinks(ctx context.Context, client *github.Client, owner, name string, issue *github.Issue) ([]RadarItem, []RadarItem, error) {
	var oldItems []RadarItem
	var newItems []RadarItem

	extractedItems, err := extractLinkedTodosFromMarkdown(issue.GetBody())
	if err != nil {
		Printf("Error parsing issue body: %#v", err)
	}
	oldItems = append(oldItems, extractedItems...)

	opts := &github.IssueListCommentsOptions{
		Sort:        github.String("created"),
		Direction:   github.String("asc"),
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		comments, resp, err := client.Issues.ListComments(ctx, owner, name, *issue.Number, opts)
		if err != nil {
			Printf("Error fetching comments: %#v", err)
			return oldItems, newItems, err
		}

		for _, comment := range comments {
			extractedItems, err := extractLinkedTodosFromMarkdown(comment.GetBody())
			if err != nil {
				Printf("Error parsing comment body: %#v", err)
			}
			newItems = append(newItems, extractedItems...)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return oldItems, newItems, nil
}

// NewGitHubClient generates a new GitHub client with the given static token source.
func NewGitHubClient(githubToken string) *github.Client {
	if _, ok := clients[githubToken]; !ok {
		clients[githubToken] = github.NewClient(oauth2.NewClient(
			context.TODO(),
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: githubToken},
			),
		))
	}

	return clients[githubToken]
}
