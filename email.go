package radar

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"time"

	"github.com/mvdan/xurls"
)

func NewEmailHandler(database *sql.DB, allowedSenders []string, debug bool) EmailHandler {
	return EmailHandler{
		AllowedSenders: allowedSenders,
		Debug:          debug,
		RadarItems:     RadarItemsService{Database: database},
		CreateQueue:    make(chan string, 10),
	}
}

type EmailHandler struct {
	// Email addresses that must be in the "From" section of the message.
	AllowedSenders []string

	// Enable debug logging.
	Debug bool

	// RadarItem service
	RadarItems RadarItemsService

	// The queue
	CreateQueue chan string
}

func (h EmailHandler) Start() {
	for url := range h.CreateQueue {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := h.RadarItems.Create(ctx, RadarItem{URL: url}); err != nil {
			log.Printf("error saving '%s': %#v %+v", url, err, err)
		} else {
			log.Printf("saved url=%s to database", url)
		}
		cancel()
	}
}

func (h EmailHandler) Shutdown(ctx context.Context) {
	close(h.CreateQueue)
	h.RadarItems.Shutdown(ctx)
}

func (h EmailHandler) IsAllowedSender(sender string) bool {
	email, err := mail.ParseAddress(sender)
	if err != nil {
		log.Printf("could not process sender '%s': %#v", sender, err)
		return false
	}

	for _, allowedSender := range h.AllowedSenders {
		if allowedSender == email.Address {
			return true
		}
	}

	return false
}

func (h EmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/x-www-form-urlencoded" {
		log.Println("don't know how to handle Content-Type:", contentType)
		http.Error(w, "cannot process Content-Type: "+contentType, http.StatusBadRequest)
		return
	}

	if sender := r.FormValue("From"); !h.IsAllowedSender(sender) {
		log.Println("not an allowed sender: ", sender)
		http.Error(w, "not an allowed sender: "+sender, http.StatusUnauthorized)
		return
	}

	emailBody := r.FormValue("body-plain")
	if h.Debug {
		log.Printf("body-plain: %#v", emailBody)
	}

	var urls []string
	if matches := xurls.Strict.FindAllString(emailBody, -1); matches != nil && len(matches) > 0 {
		urls = append(urls, matches...)
	}

	if len(urls) == 0 {
		log.Println("no urls in body: ", emailBody)
		http.Error(w, "no urls present in email body", http.StatusOK)
		return
	}

	if h.Debug {
		log.Printf("urls: %#v", urls)
	}

	for _, url := range urls {
		h.CreateQueue <- url
	}

	http.Error(w, fmt.Sprintf("added %d urls to today's radar", len(urls)), http.StatusCreated)
}
