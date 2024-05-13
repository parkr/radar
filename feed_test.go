package radar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v53/github"
	"github.com/stretchr/testify/assert"
)

func TestFeedHandler_ServeHTTP(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/search/issues", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(&github.IssuesSearchResult{
			Total:             github.Int(1),
			IncompleteResults: github.Bool(false),
			Issues: []*github.Issue{
				{
					Title:   github.String("Issue 123"),
					HTMLURL: github.String("http://example.com/issue/123"),
					Number:  github.Int(123),
					Body:    github.String(testData.newStyleBody),
				},
			},
		})
	})
	mux.HandleFunc("/repos/monalisa/diary/issues/123/comments", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]*github.IssueComment{
			{Body: github.String("- [ ] [Foo](http://bar.com)")},
		})
	})
	server := httptest.NewServer(mux)
	serverURL, _ := url.Parse(server.URL + "/")
	defer server.Close()

	ghClient := github.NewClient(nil)
	ghClient.BaseURL = serverURL
	radarItemsService := NewRadarItemsService(ghClient, "monalisa", "diary")
	feedConfig := FeedConfig{
		Title:       "My Feed",
		URL:         "http://example.com/feed.atom",
		Description: "My feed description",
		AuthorName:  "Me",
		APIKey:      "foo",
	}
	radarGeneratedChan := make(chan bool)
	h := NewFeedHandler(radarItemsService, feedConfig, radarGeneratedChan)

	req, _ := http.NewRequest("GET", "/feed.atom", nil)
	req.Form = url.Values{"tok": []string{feedConfig.APIKey}}
	recorder := httptest.NewRecorder()
	h.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "application/atom+xml", recorder.Header().Get("Content-Type"))

	body := recorder.Body.String()
	assert.NotEqual(t, "", body)
	assert.Contains(t, body, feedConfig.Title)
	assert.Contains(t, body, "Song Notes") // random item from testData
	fmt.Println(body)
}
