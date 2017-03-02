package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/parkr/radar"
)

func main() {
	var binding string
	flag.StringVar(&binding, "http", ":8291", "The IP/PORT to bind this server to.")
	flag.Parse()

	emailHandler := radar.EmailHandler{
		AllowedSenders: strings.Split(os.Getenv("RADAR_ALLOWED_SENDERS"), ","),
		Debug:          (os.Getenv("DEBUG") != ""),
	}
	http.Handle("/emails", emailHandler)
	http.Handle("/email", emailHandler)

	log.Println("Starting server on", binding)
	log.Println(http.ListenAndServe(binding, nil))
}
