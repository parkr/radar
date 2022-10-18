package radar

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"text/template"
	"time"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

// Generate and re-use one client per token. Key = token, value = client for token.
var clients = map[string]*github.Client{}

var labels = []string{"radar"}

var bodyTmpl = template.Must(template.New("body").Parse(`
{{with .OldIssueURL}}[*Previously:*]({{.}}){{end}}

{{range .OldIssues}}- [ ] [{{.GetTitle}}]({{.URL}})
{{end}}
{{with .NewIssues}}New:

{{range .}}- [ ] [{{.GetTitle}}]({{.URL}})
{{end}}{{end}}
{{with .Mention}}/cc {{.}}{{end}}
`))

type tmplData struct {
	OldIssueURL string
	NewIssues   []RadarItem
	OldIssues   []RadarItem
	Mention     string
}

func GenerateRadarIssue(radarItemsService RadarItemsService, mention string) (*github.Issue, error) {
	client := radarItemsService.githubClient
	owner, name := radarItemsService.owner, radarItemsService.repoName

	data := &tmplData{
		Mention: mention,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	previousIssue := getPreviousRadarIssue(ctx, client, owner, name)
	if previousIssue != nil {
		data.OldIssueURL = *previousIssue.HTMLURL
		data.OldIssues = extractGitHubLinks(ctx, client, owner, name, previousIssue)
	}

	sort.Stable(RadarItems(data.NewIssues))
	sort.Stable(RadarItems(data.OldIssues))

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
	if len(data.NewIssues) == 0 && len(data.OldIssues) == 0 {
		return "Nothing to do today. Nice work! :sparkles:", nil
	}

	buf := bytes.NewBufferString("A new day! Here's what you have saved:\n")
	err := bodyTmpl.Execute(buf, data)
	return buf.String(), err
}

func extractGitHubLinks(ctx context.Context, client *github.Client, owner, name string, issue *github.Issue) []RadarItem {
	var items []RadarItem

	items = append(items, extractLinkedTodosFromMarkdown(issue.GetBody())...)

	opts := &github.IssueListCommentsOptions{
		Sort:        github.String("created"),
		Direction:   github.String("asc"),
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		comments, resp, err := client.Issues.ListComments(ctx, owner, name, *issue.Number, opts)
		if err != nil {
			Printf("Error fetching comments: %#v", err)
			return items
		}

		for _, comment := range comments {
			items = append(items, extractLinkedTodosFromMarkdown(comment.GetBody())...)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	return items
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
