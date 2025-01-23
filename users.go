package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type userPost struct {
		Email string `json:"email"`
	}

	type jsonUser struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	userData := userPost{}
	err := decoder.Decode(&userData)
	if err != nil {
		log.Printf("There was an error decoding the request body: %s", err)
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), userData.Email)
	if err != nil {
		log.Printf("Error creating user: %s", err)
		return
	}
	payload := jsonUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	response, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error encoding json: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(response)
}

func (cfg *apiConfig) handleResetUsers(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.WriteHeader(403)
		return
	}
	_, err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		w.WriteHeader(500)
		return
	}
	return
}
