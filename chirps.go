package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"gitea.rannes.dev/christian/chirpy/internal/database"
	"github.com/google/uuid"
)

type cleanedMsg struct {
	CleanedMsg string `json:"cleaned_body"`
}

type chirpSelect struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpId"))
	if err != nil {
		log.Print(err)
		respondWithError(w, 400, "You must enter a valid UUID")
	}
	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, 404, fmt.Sprintf("Chirp with id %s doesn not exist", id))
	}
	c := chirpSelect{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}
	writeResponse(w, 200, c)
}

func (cfg *apiConfig) handleGetChirpList(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.ListChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, fmt.Sprintf("There was an error fetching chirls: %s", err))
	}
	if len(chirps) == 0 {
		respondWithError(w, 404, "No chirps found")
	}
	chirpList := []chirpSelect{}
	for _, chirp := range chirps {
		c := chirpSelect{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		chirpList = append(chirpList, c)
	}
	writeChirpListResponse(w, 200, chirpList)
}

func writeChirpListResponse(w http.ResponseWriter, status int, responseBody []chirpSelect) {
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

func (cfg *apiConfig) handleCreateChirp(w http.ResponseWriter, r *http.Request) {
	type chirpInsert struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	decoder := json.NewDecoder(r.Body)
	payload := chirpInsert{}
	err := decoder.Decode(&payload)
	if err != nil {
		respondWithError(w, 500, "Error decoding message")
		return
	}
	if len(payload.Body) > 140 {
		respondWithError(w, 400, "chirp too long")
		return
	}
	cMsg := censorProfanity(payload.Body)
	newChirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: cMsg, UserID: payload.UserID})
	if err != nil {
		log.Printf("There was an error saving your chirp to the db: %s", err)
	}
	writeResponse(w, 201, chirpSelect{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID,
	})
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
