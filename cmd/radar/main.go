package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/parkr/radar"
)

func main() {
	var binding string
	flag.StringVar(&binding, "http", ":8291", "The IP/PORT to bind this server to.")
	flag.Parse()

	db, err := sql.Open("mysql", os.Getenv("RADAR_MYSQL_URL"))
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}

	emailHandler := radar.EmailHandler{
		AllowedSenders: strings.Split(os.Getenv("RADAR_ALLOWED_SENDERS"), ","),
		Debug:          (os.Getenv("DEBUG") != ""),
		Database:       db,
	}
	http.Handle("/emails", handlers.LoggingHandler(os.Stdout, emailHandler))
	http.Handle("/email", handlers.LoggingHandler(os.Stdout, emailHandler))

	log.Println("Starting server on", binding)
	log.Println(http.ListenAndServe(binding, nil))
}
