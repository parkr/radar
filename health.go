package radar

import (
	"fmt"
	"net/http"

	"github.com/technoweenie/grohl"
)

type healthHandler struct {
	svc RadarItemsService
}

func (h healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, `{"ok":true}`)
	logCtx := extractLogCtx(r)
	logCtx.Log(grohl.Data{"ok": true})
}

// NewHealthHandler returns a handler which provides health-related information.
func NewHealthHandler(svc RadarItemsService) http.Handler {
	return healthHandler{svc: svc}
}
