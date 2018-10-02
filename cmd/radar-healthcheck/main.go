package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/parkr/radar"
	"github.com/technoweenie/grohl"
)

func fail(logCtx *grohl.Context, err error, format string, args ...interface{}) {
	logCtx.Log(grohl.Data{
		"msg":   fmt.Sprintf(format, args...),
		"error": err,
	})
	os.Exit(1)
}

func main() {
	healthcheckURL := os.Getenv("RADAR_HEALTHCHECK_URL")
	if healthcheckURL == "" {
		// The default here
		healthcheckURL = "http://localhost:8291/health"
	}

	logCtx := grohl.NewContext(grohl.Data{"url": healthcheckURL})

	healthCheckBody := &radar.HealthResponse{}
	resp, err := http.Get(healthcheckURL)
	if err != nil {
		fail(logCtx, err, "unable to run check health")
	}
	if err := json.NewDecoder(resp.Body).Decode(&healthCheckBody); err != nil {
		fail(logCtx, err, "unable to decode json response")
	}

	logCtx.Log(healthCheckBody.ToGrohlData())

	if !healthCheckBody.Ok {
		fail(logCtx, nil, "oh snap, not ok")
	}
}
