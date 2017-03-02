package radar

import (
	"encoding/json"
	"log"
	"net/http"
)

type EmailHandler struct {
	Debug bool
}

func (h EmailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var payload map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(payload)
	if err != nil {
		log.Println("error processing payload", err)
		http.Error(w, "error processing payload "+err.Error(), http.StatusInternalServerError)
		return
	}

	if h.Debug {
		log.Printf("%#v", payload)
	}

	http.Error(w, "hehe", http.StatusOK)
}
