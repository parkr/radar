package radar

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mvdan/xurls"
)

type EmailHandler struct {
	Debug bool
}

func (h EmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if contentType := r.Header.Get("Content-Type"); contentType != "application/x-www-form-urlencoded" {
		log.Println("Don't know how to handle Content-Type:", contentType)
		http.Error(w, "cannot process Content-Type: "+contentType, http.StatusBadRequest)
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
