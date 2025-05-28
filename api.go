package radar

import (
	"encoding/json"
	"net/http"

	"github.com/technoweenie/grohl"
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

	// Channel to signal that the radar state has changed.
	radarGeneratedChan chan bool
}

func (h APIHandler) Error(w http.ResponseWriter, message string, code int) {
	grohl.Log(grohl.Data{"status": code, "message": message})
	http.Error(w, message, code)
}

func (h APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && r.URL.Path == apiPrefix {
		h.CreateRadarItem(w, r)
		h.radarGeneratedChan <- true
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
	oldRadarItems, newRadarItems, err := h.RadarItems.List(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(map[string]interface{}{"OldItems": oldRadarItems, "NewRadarItems": newRadarItems})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
