package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	mailgun "github.com/mailgun/mailgun-go"
	"github.com/parkr/radar"
	"github.com/technoweenie/grohl"
)

// getMailgunService creates a Mailgun service from the environment variables.
func getMailgunService() radar.MailgunService {
	mg, err := mailgun.NewMailgunFromEnv()
	if err != nil {
		radar.Println("unable to fetch mailgun from env:", err)
	}
	return radar.NewMailgunService(mg, os.Getenv("MG_FROM_EMAIL"))
}

// radarGenerator handles the signals and filters so only triggers at the given hour of day generates a new radar issue.
func radarGenerator(radarItemsService radar.RadarItemsService, trigger chan os.Signal, hourToGenerateRadar string) {
	if len(hourToGenerateRadar) != 2 {
		radar.Printf("NOT generating radar. Hour to generate is not in 24-hr time: '%s'", hourToGenerateRadar)
		return
	}

	mention := os.Getenv("RADAR_MENTION")
	if mention == "" {
		radar.Println("RADAR_MENTION is empty. Just so you know.")
	}

	radar.Printf("Will generate radar at %s:00 every day.", hourToGenerateRadar)

	for signal := range trigger {
		thisHour := time.Now().Format("15")
		if thisHour == hourToGenerateRadar || signal == syscall.SIGUSR2 {
			radar.Println("The time has come: let's generate the radar!")
			generateRadar(radarItemsService, mention)
		} else {
			radar.Printf("Wrong hour to generate! %s != %s", thisHour, hourToGenerateRadar)
		}
	}
}

// generateRadar generates a new radar issue and logs it, or any errors.
func generateRadar(radarItemsService radar.RadarItemsService, mention string) {
	issue, err := radar.GenerateRadarIssue(radarItemsService, mention)
	if err == nil {
		radar.Printf("Generated new radar issue: %s", *issue.HTMLURL)
	} else {
		radar.Printf("Couldn't generate new radar issue: %#v", err)
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

	grohl.SetLogger(grohl.NewIoLogger(os.Stderr))
	grohl.SetStatter(nil, 0, "")

	mux := http.NewServeMux()

	githubToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if githubToken == "" {
		radar.Println("NOT generating radar. GITHUB_ACCESS_TOKEN not set.")
		return
	}
	radarRepo := os.Getenv("RADAR_REPO")
	if radarRepo == "" {
		radar.Println("NOT generating radar. RADAR_REPO not set.")
		return
	}

	radarRepoPieces := strings.Split(radarRepo, "/")
	radarItemsService := radar.NewRadarItemsService(radar.NewGitHubClient(githubToken), radarRepoPieces[0], radarRepoPieces[1])

	emailHandler := radar.NewEmailHandler(
		radarItemsService, // RadarItemsService
		getMailgunService(),
		strings.Split(os.Getenv("RADAR_ALLOWED_SENDERS"), ","), // Allowed senders (email addresses)
		debug, // Whether in debug mode
	)
	mux.Handle("/emails", emailHandler)
	mux.Handle("/email", emailHandler)

	apiHandler := radar.NewAPIHandler(radarItemsService, debug)
	mux.Handle("/api/", apiHandler)

	mux.Handle("/health", radar.NewHealthHandler(radarItemsService))

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

	radar.Println("Starting server on", binding)
	server := &http.Server{Addr: binding, Handler: radar.LoggingHandler(mux)}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		sig := <-c
		// sig is a ^C, handle it
		radar.Printf("Received signal %#v!", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		close(radarC)
		ticker.Stop()
		radar.Println("Shutting down radar items service...")
		radarItemsService.Shutdown(ctx)
		emailHandler.Shutdown(ctx)
		radar.Println("Telling server to shutdown...")
		_ = server.Shutdown(ctx)
		radar.Println("Done with graceful shutdown.")
	}()

	if err := server.ListenAndServe(); err != nil {
		radar.Println("error listening:", err)
	}
}
