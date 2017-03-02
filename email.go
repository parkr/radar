package radar

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/mail"

	"github.com/mvdan/xurls"
)

type EmailHandler struct {
	// Email addresses that must be in the "From" section of the message.
	AllowedSenders []string

	// Enable debug logging.
	Debug bool

	// Database to use as backend.
	Database *sql.DB
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
		log.Println("Don't know how to handle Content-Type:", contentType)
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

	http.Error(w, fmt.Sprintf("added %d urls to today's radar", len(urls)), http.StatusCreated)
}
