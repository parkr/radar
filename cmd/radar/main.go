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

	emailHandler := radar.EmailHandler{}
	http.Handle("/emails", emailHandler)
	http.Handle("/email", emailHandler)
	fmt.Println(http.ListenAndServe(binding, nil))
}
