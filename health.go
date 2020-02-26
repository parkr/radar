package radar

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/technoweenie/grohl"
)

type healthHandler struct {
	svc RadarItemsService
}

// HealthResponse is the struct representing the JSON returned from the /health endpoint.
type HealthResponse struct {
	Ok bool
	DB bool
}

// ToGrohlData returns grohl data for this health response.
func (r HealthResponse) ToGrohlData() grohl.Data {
	return grohl.Data{
		"ok": r.Ok,
		"db": r.DB,
	}
}

func newHealthResponse(ctx context.Context, db *sql.DB) HealthResponse {
	if db == nil {
		return HealthResponse{
			Ok: false,
			DB: false,
		}
	}

	err := db.PingContext(ctx)
	return HealthResponse{
		Ok: err == nil,
		DB: err == nil,
	}
}

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := newHealthResponse(r.Context(), h.svc.Database)
	if !resp.Ok {
		w.WriteHeader(http.StatusBadGateway)
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	logCtx := GetLogContext(r)
	logCtx.Log(resp.ToGrohlData())
}

// NewHealthHandler returns a handler which provides health-related information.
func NewHealthHandler(svc RadarItemsService) http.Handler {
	return healthHandler{svc: svc}
}
