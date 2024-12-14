package validation

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

type chirp struct {
	Body string `json:"body"`
}

type cleanedMsg struct {
	CleanedMsg string `json:"cleaned_body"`
}

func ValidateChirp(w http.ResponseWriter, r *http.Request) {
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
	cMsg := censorProfanity(body.Body)
	writeResponse(w, 200, cleanedMsg{
		CleanedMsg: cMsg,
	})
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

func censorProfanity(msg string) string {
	profList := []string{"kerfuffle", "sharbert", "fornax"}
	msgSlice := strings.Split(msg, " ")
	for i, v := range msgSlice {
		if slices.Contains(profList, strings.ToLower(v)) {
			msgSlice = slices.Replace(msgSlice, i, i+1, "****")
		}
	}
	censoredMsg := strings.Join(msgSlice, " ")
	return censoredMsg
}
