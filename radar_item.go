package radar

import (
	"context"
	"fmt"
	"net/url"

	"github.com/google/go-github/v48/github"
	"github.com/pkg/errors"
)

// RadarItem is a single link in the radar.
//
// RadarItem.GetTitle() is defined in parser.go. Use that to fetch the title!
type RadarItem struct {
	ID    int64
	URL   string
	Title string

	parsedURL *url.URL
}

func (r *RadarItem) GetHostname() string {
	if r.parsedURL == nil {
		var err error
		r.parsedURL, err = url.Parse(r.URL)
		if err != nil {
			Printf("GetHostname: couldn't parse URL %q: %+v", r.URL, err)
			return ""
		}
	}

	return r.parsedURL.Hostname()
}

func (r *RadarItem) GetMarkdown() string {
	return "[" + r.GetTitle() + "](" + r.URL + ")"
}

type RadarItems []RadarItem

func (r RadarItems) Len() int {
	return len(r)
}

func (r RadarItems) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RadarItems) Less(i, j int) bool {
	return r[i].GetHostname() < r[j].GetHostname()
}

// RadarItemsService can be used to fetch the radar issue, list radar items, and add a new radar item.
type RadarItemsService struct {
	githubClient *github.Client
	owner        string
	repoName     string
}

// NewRadarItemsService creates a new RadarItemsService with all the proper fields initialized.
func NewRadarItemsService(githubClient *github.Client, owner, repoName string) RadarItemsService {
	return RadarItemsService{
		githubClient: githubClient,
		owner:        owner,
		repoName:     repoName,
	}
}

// GetGitHubIssue fetches the GitHub issue.
func (rs RadarItemsService) GetGitHubIssue(ctx context.Context) (*github.Issue, error) {
	issue := getPreviousRadarIssue(ctx, rs.githubClient, rs.owner, rs.repoName)
	if issue == nil {
		newIssue, _, err := rs.githubClient.Issues.Create(ctx, rs.owner, rs.repoName, &github.IssueRequest{
			Title:  github.String(getTitle()),
			Body:   github.String("Welcome to your new radar!"),
			Labels: &labels,
		})
		if err != nil {
			return nil, err
		}
		issue = newIssue
	}
	return issue, nil
}

// List returns a list of parsed radar items present on the list.
func (rs RadarItemsService) List(ctx context.Context) ([]RadarItem, error) {
	issue, err := rs.GetGitHubIssue(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "error fetching open issue")
	}
	return extractGitHubLinks(ctx, rs.githubClient, rs.owner, rs.repoName, issue), nil
}

// Create adds a RadarItem to the GitHub issue.
func (rs RadarItemsService) Create(ctx context.Context, m RadarItem) error {
	issue, err := rs.GetGitHubIssue(ctx)
	if err != nil {
		return errors.WithMessage(err, "error fetching open issue")
	}
	_, _, err = rs.githubClient.Issues.CreateComment(ctx, rs.owner, rs.repoName, *issue.Number, &github.IssueComment{
		Body: github.String(fmt.Sprintf("- [ ] [%s](%s)", m.GetTitle(), m.URL)),
	})
	return err
}

// Shutdown closes the database connection.
func (rs RadarItemsService) Shutdown(ctx context.Context) {
}
