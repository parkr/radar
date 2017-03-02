package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/parkr/radar"
)

func getDB() *sql.DB {
	db, err := sql.Open("mysql", os.Getenv("RADAR_MYSQL_URL"))
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	return db
}

func main() {
	var binding string
	flag.StringVar(&binding, "http", ":8291", "The IP/PORT to bind this server to.")
	flag.Parse()

	emailHandler := radar.EmailHandler{
		AllowedSenders: strings.Split(os.Getenv("RADAR_ALLOWED_SENDERS"), ","),
		Debug:          (os.Getenv("DEBUG") != ""),
		Database:       getDB(),
	}
	http.Handle("/emails", handlers.LoggingHandler(os.Stdout, emailHandler))
	http.Handle("/email", handlers.LoggingHandler(os.Stdout, emailHandler))

	log.Println("Starting server on", binding)
	server := &http.Server{Addr: binding}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Printf("Received signal %#v!", sig)
			log.Println("Closing database connection...")
			emailHandler.Database.Close()
			log.Println("Telling server to shutdown...")
			server.Shutdown(context.Background())
			log.Println("Done with graceful shutdown.")
		}
	}()

	log.Println(server.ListenAndServe())
}
