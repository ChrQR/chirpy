package validation

import (
	"encoding/json"
	"log"
	"net/http"
)

func ValidateChirp(w http.ResponseWriter, r *http.Request) {
	type chirp struct {
		Body string `json:"body"`
	}
	decoder := json.NewDecoder(r.Body)
	body := chirp{}
	err := decoder.Decode(&body)
	if err != nil {
		respondWithError(w, 500, "Error decoding message")
	}
	if len(body.Body) > 140 {
		respondWithError(w, 400, "chirp too long")
		return
	}
	validResponse(w)
}

func validResponse(w http.ResponseWriter) {
	type validMsg struct {
		Valid bool `json:"valid"`
	}
	respBody := validMsg{
		Valid: true,
	}
	writeResponse(w, 200, respBody)
}

func respondWithError(w http.ResponseWriter, status int, msg string) {
	type errorMsg struct {
		Error string `json:"error"`
	}
	respBody := errorMsg{
		Error: msg,
	}
	writeResponse(w, status, respBody)
}

func writeResponse(w http.ResponseWriter, status int, responseBody interface{}) {
	body, err := json.Marshal(responseBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if status != 200 {
		w.WriteHeader(status)
	}
	w.Write(body)
}
