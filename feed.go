package radar

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/feeds"
)

type FeedConfig struct {
	Title       string
	URL         string
	Description string
	AuthorName  string
	APIKey      string
}

type FeedHandler struct {
	radarItems         RadarItemsService
	feed               *feeds.Feed
	apiToken           string
	cache              bytes.Buffer
	radarGeneratedChan chan bool
}

// NewFeedHandler creates a new handler which will respond with an Atom feed of radar items.
func NewFeedHandler(radarItemsService RadarItemsService, config FeedConfig, radarGeneratedChan chan bool) *FeedHandler {
	return &FeedHandler{
		radarItems: radarItemsService,
		feed: &feeds.Feed{
			Title:       config.Title,
			Link:        &feeds.Link{Href: config.URL},
			Description: config.Description,
			Author:      &feeds.Author{Name: config.AuthorName},
			Created:     time.Now(),
		},
		radarGeneratedChan: radarGeneratedChan,
		apiToken:           config.APIKey,
	}
}

func (h FeedHandler) Start() {
	for _ = range h.radarGeneratedChan {
		h.ResetCache()
	}
}

func convertRadarItemToFeedItem(item RadarItem) *feeds.Item {
	return &feeds.Item{
		Title:       item.GetTitle(),
		Link:        &feeds.Link{Href: item.URL},
		Description: item.GetHostname(),
		Content:     item.GetFormatted(),
		Created:     time.Now(),
	}
}

func (h FeedHandler) ResetCache() {
	h.cache.Reset()
}

func (h *FeedHandler) populateCache(ctx context.Context) error {
	h.ResetCache()

	oldItems, newItems, err := h.radarItems.List(ctx)
	if err != nil {
		return err
	}

	h.feed.Items = []*feeds.Item{}

	for _, item := range newItems {
		h.feed.Items = append(h.feed.Items, convertRadarItemToFeedItem(item))
	}

	for _, item := range oldItems {
		h.feed.Items = append(h.feed.Items, convertRadarItemToFeedItem(item))
	}

	return h.feed.WriteAtom(&h.cache)
}

func (h FeedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("tok") != h.apiToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := h.populateCache(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/atom+xml")
	if _, err := io.Copy(w, &h.cache); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
