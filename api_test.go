package radar

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-github/v53/github"
	"github.com/stretchr/testify/assert"
)

func TestApiHandler_UnsupportedMethod(t *testing.T) {
	// Create a new APIHandler with a mock RadarItemsService
	mockService := RadarItemsService{}
	handler := NewAPIHandler(mockService, false, make(chan bool, 100))

	// Create a new HTTP request with an unsupported method
	req, err := http.NewRequest("PUT", apiPrefix, nil)
	assert.NoError(t, err, "Failed to create request")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code, "Expected status code 404 Not Found")
	expectedBody := "404 not found at all\n"
	assert.Equal(t, expectedBody, rr.Body.String(), "Expected response body to match")
}

func TestApiHandler_ListItems(t *testing.T) {
	// Create a new APIHandler with a mock RadarItemsService
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
			{Body: github.String(testData.comment1Body)},
			{Body: github.String(testData.comment2Body)},
		})
	})
	server := httptest.NewServer(mux)
	serverURL, _ := url.Parse(server.URL + "/")
	defer server.Close()

	ghClient := github.NewClient(nil)
	ghClient.BaseURL = serverURL
	radarItemsService := NewRadarItemsService(ghClient, "monalisa", "diary")
	debug := false
	handler := NewAPIHandler(radarItemsService, debug, make(chan bool, 100))

	// Create a new HTTP request for listing items
	req, err := http.NewRequest("GET", apiPrefix, nil)
	assert.NoError(t, err, "Failed to create request")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code 200 OK")
	// Parse the response body into apiListItemsResponse
	var response apiListItemsResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal response body")
	assert.Len(t, response.OldRadarItems, len(testData.expectedOldItems), "Expected all test data old items")
	assert.Len(t, response.NewRadarItems, len(testData.expectedNewItems), "Expected all test data new items")
}

func TestApiHandler_CreateItem(t *testing.T) {
	// Create a new APIHandler with a mock RadarItemsService
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
		assert.Equal(t, http.MethodPost, r.Method)
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err, "Failed to read request body")
		defer r.Body.Close()
		inputIssueComment := &github.IssueComment{}
		err = json.Unmarshal(body, inputIssueComment)
		assert.NoError(t, err, "Failed to unmarshal request body into IssueComment")
		assert.Equal(t, "- [ ] [Some Great Site](https://somegreat.site)", inputIssueComment.GetBody())

		w.WriteHeader(http.StatusCreated) // Respond with 201 Created
		_, err = w.Write(body)            // Echo back the request body
		assert.NoError(t, err, "Failed to write response body")
	})
	server := httptest.NewServer(mux)
	serverURL, _ := url.Parse(server.URL + "/")
	defer server.Close()

	ghClient := github.NewClient(nil)
	ghClient.BaseURL = serverURL
	radarItemsService := NewRadarItemsService(ghClient, "monalisa", "diary")
	debug := false
	handler := NewAPIHandler(radarItemsService, debug, make(chan bool, 100))

	// Create a new HTTP request for listing items
	form := url.Values{}
	form.Set("url", "https://somegreat.site")
	form.Set("title", "Some Great Site")
	req, err := http.NewRequest("POST", apiPrefix, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	assert.NoError(t, err, "Failed to create request")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Expected status code 201 Created")
}
