package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/parkr/radar"
	"github.com/technoweenie/grohl"
)

// indexHtml is the HTML for the index page that allows users to create a new radar item.
var indexHtml = `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Create Radar Item</title>
	<link rel="stylesheet" href="/styles.css" />
	<script type="text/javascript">
	function clearForm() {
		document.getElementById("title").value = "";
		document.getElementById("url").value = "";
	}
	</script>
</head>
<body>
	<h1>Create Radar Item</h1>
	<form action="/create" method="post">
		<label for="title">Title:</label>
		<input type="text" id="title" name="title">
		<label for="url">URL:</label>
		<input type="url" id="url" name="url" required>
		<button type="submit">Create</button>
		<button onclick="clearForm()">Clear</button>
	</form>
	<p>Use this form to create a new radar item. The title is optional, but the URL is required.</p>
	<p>After submitting, the radar item will be created and you will receive a confirmation message.</p>
</body>
</html>
`

// createHTML is the HTML for the create confirmation page after a radar item is created.
var createHTMLTemplate = template.Must(template.New("createHTML").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Radar Item Created</title>
	<link rel="stylesheet" href="/styles.css" />
</head>
<body>
	<h1>Radar Item Created</h1>
	<p>{{.ConfirmationMessage}}</p>
	<p><strong>Title:</strong> {{.Title}}</p>
	<p><strong>URL:</strong> <a href="{{.URL}}" target="_blank">{{.URL}}</a></p>
	<p>Thank you for contributing to the radar!</p>
	<p><a href="/">Go back to the form</a></p>
</body>
</html>
`))

const stylesCSS = `
body {
		font-family: Arial, sans-serif;
		background-color: #f4f4f4;
		color: #333;
		margin: 0;
		padding: 20px;
	}
	h1 {
		color: #444;
	}
	form {
		background-color: #fff;
		padding: 20px;
		border-radius: 5px;
		box-shadow: 0 2px 5px rgba(0, 0, 0, 0.1);
	}
	label {
		display: block;
		margin-bottom: 10px;
	}
	input[type="text"],
	input[type="url"] {
		width: 100%;
		padding: 10px;
		margin-bottom: 10px;
		border: 1px solid #ccc;
		border-radius: 4px;
	}
	input[type="text"]:focus,
	input[type="url"]:focus {
		border-color: #007bff;
		outline: none;
		box-shadow: 0 0 5px rgba(0, 123, 255, 0.5);
	}
	button {
		background-color: #007bff;
		color: white;
		border: none;
		padding: 10px 15px;
		border-radius: 4px;
		cursor: pointer;
	}
	button:hover {
		background-color: #0056b3;
	}
	p {
		margin-top: 20px;
		font-size: 0.9em;
		color: #666;
	}
	@media(prefers-color-scheme: dark) {
		body {
			background-color: #333;
			color: #f4f4f4;
		}
		h1 {
			color: #f4f4f4;
		}
		form {
			background-color: #444;
			color: #f4f4f4;
			box-shadow: 0 2px 5px rgba(0, 0, 0, 0.3);
		}
		input[type="text"],
		input[type="url"] {
			background-color: #555;
			color: #f4f4f4;
			border: 1px solid #666;
		}
		input[type="text"]:focus,
		input[type="url"]:focus {
			border-color: #007bff;
			box-shadow: 0 0 5px rgba(0, 123, 255, 0.5);
		}
		button {
			background-color: #007bff;
			color: white;
			border: none;
			padding: 10px 15px;
			border-radius: 4px;
			cursor: pointer;
		}
		button:hover {
			background-color: #0056b3;
		}
		p {
			color: #ccc;
		}
	}
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
	_, _ = w.Write([]byte(indexHtml))
}

func styles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(stylesCSS))
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
			renderCreateHTML(
				w,
				http.StatusInternalServerError,
				"", "",
				"Failed to create radar item due to an internal error.")
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
			renderCreateHTML(
				w,
				http.StatusInternalServerError,
				"", "",
				"Failed to create radar item due to an internal error.")
			return
		}
		grohl.Log(grohl.Data{"msg": "received response from radar items service", "status": resp.Status})
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			grohl.Log(grohl.Data{"msg": "error creating radar item", "status": resp.Status})
			renderCreateHTML(
				w,
				resp.StatusCode,
				"", "",
				fmt.Sprintf("Failed to create radar item due to an error with the radar service: %s.", resp.Status))
			return
		}
		grohl.Log(grohl.Data{"msg": "radar item created successfully", "status": resp.Status})
		renderCreateHTML(
			w,
			http.StatusCreated,
			r.FormValue("title"),
			r.FormValue("url"),
			"Your radar item has been created successfully! You can view it on the radar page.")
		grohl.Log(grohl.Data{"msg": "radar item created", "url": r.FormValue("url"), "title": r.FormValue("title")})
	}
}

func renderCreateHTML(out http.ResponseWriter, code int, title, url, confirmationMessage string) error {
	out.WriteHeader(code)
	data := struct {
		Title               string
		URL                 string
		ConfirmationMessage string
	}{
		Title:               title,
		URL:                 url,
		ConfirmationMessage: confirmationMessage,
	}
	if err := createHTMLTemplate.Execute(out, data); err != nil {
		grohl.Log(grohl.Data{"msg": "error rendering create HTML", "err": err})
		return err
	}
	return nil
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
		os.Exit(1)
	}
	conf.RadarItemsServiceURL = u

	grohl.SetLogger(grohl.NewIoLogger(os.Stderr))
	grohl.SetStatter(nil, 0, "")

	mux := http.NewServeMux()

	mux.Handle("POST /create", http.HandlerFunc(create(*conf)))
	mux.Handle("GET /styles.css", http.HandlerFunc(styles))
	mux.Handle("GET /", http.HandlerFunc(show))

	server := &http.Server{Addr: conf.ServeURL, Handler: radar.LoggingHandler(mux)}

	radar.Printf("Starting server on %s", conf.ServeURL)
	if err := server.ListenAndServe(); err != nil {
		radar.Println("error listening:", err)
	}
}
