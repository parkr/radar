package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/parkr/radar"
)

func main() {
	var binding string
	flag.StringVar(&binding, "http", ":8291", "The IP/PORT to bind this server to.")
	flag.Parse()

	http.Handle("/emails", radar.EmailHandler{})
	fmt.Println(http.ListenAndServe(binding, nil))
}
