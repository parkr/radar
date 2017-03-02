package radar

import (
	"io/ioutil"
	"log"
	"net/http"
)

type EmailHandler struct {
}

func (h EmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	json, err := ioutil.ReadAll(r.Body)
	log.Println(json, err)
	http.Error(w, "hehe", http.StatusOK)
}
