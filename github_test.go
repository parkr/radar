package radar

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func newTestGitHubClientAndServer(t *testing.T) (*github.Client, *httptest.Server, error) {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if r.URL.Path == `/api/v3/repos/parkr-test/radar-test/issues/1887/comments` && r.FormValue("page") == "" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Link", `<https://api.github.com/api/v3/repos/parkr-test/radar-test/issues/1887/comments?page=2>; rel="next", <https://api.github.com/api/v3/repos/parkr-test/radar-test/issues/1887/comments?page=2>; rel="last", <https://api.github.com/api/v3/repos/parkr-test/radar-test/issues/1887/comments?page=1>; rel="first"`)
			json.NewEncoder(w).Encode([]*github.IssueComment{
				{Body: github.String(testData.comment1Body)},
			})
			return
		}
		if r.URL.Path == `/api/v3/repos/parkr-test/radar-test/issues/1887/comments` && r.FormValue("page") == "2" {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Link", `<https://api.github.com/api/v3/repos/parkr-test/radar-test/issues/1887/comments?page=1>; rel="prev", <https://api.github.com/api/v3/repos/parkr-test/radar-test/issues/1887/comments?page=2>; rel="last", <https://api.github.com/api/v3/repos/parkr-test/radar-test/issues/1887/comments?page=1>; rel="first"`)
			json.NewEncoder(w).Encode([]*github.IssueComment{
				{Body: github.String(testData.comment2Body)},
			})
			return
		}
		if r.URL.Path == `/api/v3/search/issues` && r.FormValue("q") == `repo:parkr-test/radar-test is:open is:issue label:radar` {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(&github.IssuesSearchResult{
				Total:             github.Int(1),
				IncompleteResults: github.Bool(false),
				Issues: []*github.Issue{
					{
						Body:    github.String(testData.newStyleBody),
						HTMLURL: github.String("https://github.com/parkr-test/radar-test/issues/1887"),
						Number:  github.Int(1887),
					},
				},
			})
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == "/api/v3/repos/parkr-test/radar-test/issues" {
			issueCreatePayload := &github.IssueRequest{}
			assert.NoError(t, json.NewDecoder(r.Body).Decode(issueCreatePayload))
			assert.Equal(t, testData.newIssueBody, issueCreatePayload.GetBody())
			assert.Equal(t, `Radar for `+time.Now().Format("2006-01-02"), issueCreatePayload.GetTitle())

			w.Header().Set("Content-Type", "application/json")
			issueCreateResponse := &github.Issue{}
			assert.NoError(t, json.NewEncoder(w).Encode(issueCreateResponse))
			return
		}
		if r.Method == http.MethodPatch && r.URL.Path == `/api/v3/repos/parkr-test/radar-test/issues/1887` {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{}`, http.StatusOK)
			return
		}
		r.ParseForm()
		log.Println(r.Method, r.URL.Path, r.Form)
		http.Error(w, `{}`, http.StatusNotFound)
	})

	server := httptest.NewServer(handler)
	client, err := github.NewEnterpriseClient(
		server.URL,
		server.URL,
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "fakegithubauthtoken"}),
		),
	)
	return client, server, err
}

func TestJoinLinksIntoBody(t *testing.T) {
	header := "A new day, @parkr! Here's what you have saved:"
	oldLinks := []RadarItem{{URL: "https://google.com"}, {URL: "https://ben.balter.com", Title: "Ben Balter"}}
	newLinks := []RadarItem{{URL: "https://byparker.com"}, {URL: "https://jvns.ca"}}
	expected := header + `

## New:

  * [ ] [Parker Moore | By Parker](https://byparker.com)
  * [ ] [Julia Evans](https://jvns.ca)

## *Previously:*

  * [ ] [Google](https://google.com)
  * [ ] [Ben Balter](https://ben.balter.com)

*Previously:* https://github.com/parkr/radar/issues/1
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

func Test_extractGitHubLinks_newStyleBody(t *testing.T) {
	client, server, err := newTestGitHubClientAndServer(t)
	defer server.Close()
	assert.NoError(t, err)

	oldItems, newItems, err := extractGitHubLinks(context.Background(), client, "parkr-test", "radar-test", &github.Issue{
		Body:   github.String(testData.newStyleBody),
		Number: github.Int(1887),
	})

	assert.NoError(t, err)
	assert.Equal(t, testData.expectedOldItems, oldItems)
	assert.Equal(t, testData.expectedNewItems, newItems)
}

func Test_extractGitHubLinks_oldStyleBody(t *testing.T) {
	client, server, err := newTestGitHubClientAndServer(t)
	defer server.Close()
	if err != nil {
		t.Fatalf("unable to build test github client: %v", err)
	}

	oldItems, newItems, err := extractGitHubLinks(context.Background(), client, "parkr-test", "radar-test", &github.Issue{
		Body:   github.String(testData.oldStyleBody),
		Number: github.Int(1887),
	})

	assert.NoError(t, err)
	assert.Equal(t, testData.expectedOldItems, oldItems)
	assert.Equal(t, testData.expectedNewItems, newItems)
}

func TestGenerateRadarIssue(t *testing.T) {
	client, server, err := newTestGitHubClientAndServer(t)
	defer server.Close()
	assert.NoError(t, err)
	service := RadarItemsService{githubClient: client, owner: "parkr-test", repoName: "radar-test"}

	_, err = GenerateRadarIssue(service, "@monalisa")
	assert.NoError(t, err)
}
