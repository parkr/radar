package radar

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func NewAPIHandler(db *sql.DB, debug bool) APIHandler {
	return APIHandler{
		RadarItems: RadarItemsService{Database: db},
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

	if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, apiPrefix+"/") {
		h.GetRadarItem(w, r)
		return
	}

	http.Error(w, "404 not found", http.StatusNotFound)
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
	radarItems, err := h.RadarItems.List(r.Context(), -1)
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

func (h APIHandler) GetRadarItem(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, apiPrefix)
	if idStr == "" {
		h.Error(w, "must submit a numerical id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.Error(w, "not a numerical id: "+idStr, http.StatusBadRequest)
		return
	}

	radarItem, err := h.RadarItems.Get(r.Context(), int64(id))
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			h.Error(w, "no radar item with id="+idStr, http.StatusNotFound)
			return
		}
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(radarItem)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
