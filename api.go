package radar

import (
	"encoding/json"
	"log"
	"net/http"
)

func NewAPIHandler(radarItemsService RadarItemsService, debug bool) APIHandler {
	return APIHandler{
		RadarItems: radarItemsService,
		Debug:      debug,
	}
}

var apiPrefix = "/api/radar_items"

type APIHandler struct {
	// RadarItem service
	RadarItems RadarItemsService

	// Enable debug logging.
	Debug bool
}

func (h APIHandler) Error(w http.ResponseWriter, message string, code int) {
	log.Printf("status=%d message=\"%s\"", code, message)
	http.Error(w, message, code)
}

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == apiPrefix {
		h.CreateRadarItem(w, r)
		return
	}

	if r.Method == http.MethodGet && r.URL.Path == apiPrefix {
		h.ListRadarItems(w, r)
		return
	}

	h.Error(w, "404 not found at all", http.StatusNotFound)
}

func (h APIHandler) CreateRadarItem(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if url == "" {
		h.Error(w, "url cannot be blank", http.StatusBadRequest)
		return
	}

	err := h.RadarItems.Create(r.Context(), RadarItem{
		URL:   url,
		Title: r.FormValue("title"),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Error(w, "successfully saved url", http.StatusCreated)
}

func (h APIHandler) ListRadarItems(w http.ResponseWriter, r *http.Request) {
	radarItems, err := h.RadarItems.List(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(radarItems)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
