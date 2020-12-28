package radar

import (
	"context"
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
}

// ToGrohlData returns grohl data for this health response.
func (r HealthResponse) ToGrohlData() grohl.Data {
	return grohl.Data{
		"ok": r.Ok,
	}
}

func newHealthResponse(ctx context.Context) HealthResponse {
	return HealthResponse{
		Ok: true,
	}
}

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := newHealthResponse(r.Context())
	if !resp.Ok {
		w.WriteHeader(http.StatusBadGateway)
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	logCtx := GetLogContext(r)
	_ = logCtx.Log(resp.ToGrohlData())
}

// NewHealthHandler returns a handler which provides health-related information.
func NewHealthHandler(svc RadarItemsService) http.Handler {
	return healthHandler{svc: svc}
}
