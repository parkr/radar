package radar

import (
	"io/ioutil"
	"log"
	"net/http"
)

type EmailHandler struct {
	Debug bool
}

func (h EmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	log.Printf("%#v", r)

	stuff, err := ioutil.ReadAll(r.Body)
	r.Body.Close()

	if err != nil {
		log.Println("error processing payload", err)
		http.Error(w, "error processing payload "+err.Error(), http.StatusInternalServerError)
		return
	}

	payload := string(stuff)

	if h.Debug {
		log.Printf("%#v", payload)
	}

	http.Error(w, "hehe", http.StatusOK)
}
