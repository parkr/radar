package main

import (
	"flag"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/parkr/radar"
	"github.com/technoweenie/grohl"
)

var createHtmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Create Radar Item</title>
</head>
<body>
	<h1>Create Radar Item</h1>
	<form action="/create" method="post">
		<label for="title">Title:</label>
		<input type="text" id="title" name="title">
		<label for="url">URL:</label>
		<input type="url" id="url" name="url" required>
		<button type="submit">Create</button>
	</form>
</body>
</html>
`

type config struct {
	ServeURL string

	RadarItemsServiceUnparsedURL string
	RadarItemsServiceURL         *url.URL
	RadarItemsServiceToken       string
}

func show(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(createHtmlTemplate))
}

func create(conf config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form := url.Values{}
		form.Set("url", r.FormValue("url"))
		form.Set("title", r.FormValue("title"))

		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequest(http.MethodPost, conf.RadarItemsServiceURL.String(), strings.NewReader(form.Encode()))
		if err != nil {
			grohl.Log(grohl.Data{"msg": "error creating request to radar items service", "err": err})
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if conf.RadarItemsServiceToken != "" {
			req.Header.Set("Authorization", "Bearer "+conf.RadarItemsServiceToken)
		}
		grohl.Log(grohl.Data{"msg": "sending request to radar items service", "url": conf.RadarItemsServiceURL.String(), "form": form})
		resp, err := client.Do(req)
		if err != nil {
			grohl.Log(grohl.Data{"msg": "error sending request to radar items service", "err": err})
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		grohl.Log(grohl.Data{"msg": "received response from radar items service", "status": resp.Status})
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			grohl.Log(grohl.Data{"msg": "error creating radar item", "status": resp.Status})
			http.Error(w, "Failed to create radar item", resp.StatusCode)
			return
		}
		grohl.Log(grohl.Data{"msg": "radar item created successfully", "status": resp.Status})
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("Radar item created successfully.\n"))
		grohl.Log(grohl.Data{"msg": "radar item created", "url": r.FormValue("url"), "title": r.FormValue("title")})
	}
}

func main() {
	conf := &config{}

	flag.StringVar(&conf.ServeURL, "http", ":3000", "The IP/PORT to bind this server to.")
	flag.StringVar(&conf.RadarItemsServiceUnparsedURL, "url", "http://localhost:8291/api/radar_items", "The URL of the radar items service.")
	flag.StringVar(&conf.RadarItemsServiceToken, "token", "", "The token for the radar items service.")
	flag.Parse()

	u, err := url.Parse(conf.RadarItemsServiceUnparsedURL)
	if err != nil {
		grohl.Log(grohl.Data{"msg": "error parsing radar items service URL:", "err": err})
	}
	conf.RadarItemsServiceURL = u

	grohl.SetLogger(grohl.NewIoLogger(os.Stderr))
	grohl.SetStatter(nil, 0, "")

	mux := http.NewServeMux()

	mux.Handle("/", http.HandlerFunc(show))
	mux.Handle("/create", http.HandlerFunc(create(*conf)))

	server := &http.Server{Addr: conf.ServeURL, Handler: radar.LoggingHandler(mux)}

	radar.Printf("Starting server on %s", conf.ServeURL)
	if err := server.ListenAndServe(); err != nil {
		radar.Println("error listening:", err)
	}
}
