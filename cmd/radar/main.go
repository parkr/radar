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
	"syscall"
	"time"

	"github.com/go-http-utils/logger"
	_ "github.com/go-sql-driver/mysql"
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

func getRadarItemsService() radar.RadarItemsService {
	return radar.RadarItemsService{Database: getDB()}
}

func radarGenerator(radarItemsService radar.RadarItemsService, trigger chan os.Signal, hourToGenerateRadar string) {
	if len(hourToGenerateRadar) != 2 {
		log.Printf("NOT generating radar. Hour to generate is not in 24-hr time: '%s'", hourToGenerateRadar)
		return
	}

	githubToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if githubToken == "" {
		log.Println("NOT generating radar. GITHUB_ACCESS_TOKEN not set.")
		return
	}

	radarRepo := os.Getenv("RADAR_REPO")
	if githubToken == "" {
		log.Println("NOT generating radar. RADAR_REPO not set.")
		return
	}

	mention := os.Getenv("RADAR_MENTION")
	if mention == "" {
		log.Println("RADAR_MENTION is empty. Just so you know.")
	}

	log.Printf("Will generate radar at %s:00 every day.", hourToGenerateRadar)

	for signal := range trigger {
		thisHour := time.Now().Format("15")
		if thisHour == hourToGenerateRadar || signal == syscall.SIGUSR2 {
			log.Println("The time has come: let's generate the radar!")
			generateRadar(radarItemsService, githubToken, radarRepo, mention)
		} else {
			log.Printf("Wrong hour to generate! %s != %s", thisHour, hourToGenerateRadar)
		}
	}
}

func generateRadar(radarItemsService radar.RadarItemsService, githubToken, radarRepo, mention string) {
	issue, err := radar.GenerateRadarIssue(radarItemsService, githubToken, radarRepo, mention)
	if err == nil {
		log.Printf("Generated new radar issue: %s", *issue.HTMLURL)
	} else {
		log.Printf("Couldn't generate new radar issue: %#v", err)
	}
}

func main() {
	var binding string
	flag.StringVar(&binding, "http", ":8291", "The IP/PORT to bind this server to.")
	var debug bool
	flag.BoolVar(&debug, "debug", os.Getenv("DEBUG") == "", "Whether to print debugging messages.")
	var hourToGenerateRadar string
	flag.StringVar(&hourToGenerateRadar, "hour", "03", "Hour of day (01-23) to generate the radar message.")
	flag.Parse()

	mux := http.NewServeMux()
	radarItemsService := getRadarItemsService()

	emailHandler := radar.NewEmailHandler(
		radarItemsService, // RadarItemsService
		strings.Split(os.Getenv("RADAR_ALLOWED_SENDERS"), ","), // Allowed senders (email addresses)
		debug, // Whether in debug mode
	)
	mux.Handle("/emails", emailHandler)
	mux.Handle("/email", emailHandler)

	apiHandler := radar.NewAPIHandler(radarItemsService, debug)
	mux.Handle("/api/", apiHandler)

	go emailHandler.Start()

	// Start the radarGenerator.
	radarC := make(chan os.Signal, 1)
	go radarGenerator(radarItemsService, radarC, hourToGenerateRadar)

	// Sending SIGUSR2 to this process generates a radar.
	signal.Notify(radarC, syscall.SIGUSR2)

	// Prompt radarGenerator to do something every 1 hour.
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			radarC <- syscall.SIGUSR1
		}
	}()

	log.Println("Starting server on", binding)
	server := &http.Server{Addr: binding, Handler: logger.Handler(mux, os.Stderr, logger.TinyLoggerType)}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			log.Printf("Received signal %#v!", sig)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			close(radarC)
			ticker.Stop()
			log.Println("Closing database connection...")
			radarItemsService.Shutdown(ctx)
			emailHandler.Shutdown(ctx)
			log.Println("Telling server to shutdown...")
			server.Shutdown(ctx)
			log.Println("Done with graceful shutdown.")
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Println("error listening:", err)
	}
}
