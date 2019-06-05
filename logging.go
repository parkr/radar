package radar

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/technoweenie/grohl"
)

type logCtxKeyValue struct{}

// For some reason, Go doesn't like a basic string value here.
// Do something TOTALLY WILD instead.
var logCtxKey = &logCtxKeyValue{}

type loggingHandler struct {
	handler http.Handler
}

func (h loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logCtx := grohl.NewContext(grohl.Data{
		"url":        r.URL.Path,
		"request_id": uuid.New().String(),
	})
	logCtx.SetStatter(nil, 0, "")
	timer := logCtx.Timer(grohl.Data{})
	h.handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), logCtxKey, logCtx)))
	timer.Finish()
}

// LoggingHandler logs pertinent request information.
func LoggingHandler(handler http.Handler) http.Handler {
	return loggingHandler{handler: handler}
}

func extractLogCtx(req *http.Request) *grohl.Context {
	return req.Context().Value(logCtxKey).(*grohl.Context)
}

// Printf prints the input using grohl.
func Printf(format string, args ...interface{}) {
	grohl.Log(grohl.Data{"msg": fmt.Sprintf(format, args...)})
}

// Println prints the input using grohl.
func Println(args ...interface{}) {
	grohl.Log(grohl.Data{"msg": strings.TrimSuffix(fmt.Sprintln(args...), "\n")})
}
