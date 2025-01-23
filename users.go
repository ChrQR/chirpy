package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"gitea.rannes.dev/christian/chirpy/internal/auth"
	"gitea.rannes.dev/christian/chirpy/internal/database"
	"github.com/google/uuid"
)

type PostUser struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JsonUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	userData := PostUser{}
	err := decoder.Decode(&userData)
	if err != nil {
		log.Printf("There was an error decoding the request body: %s", err)
		return
	}
	hashed, err := auth.HashPassword(userData.Password)
	if err != nil {
		respondWithError(w, 500, "There was an error hashing your password")
		return
	}
	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          userData.Email,
		HashedPassword: hashed,
	})
	if err != nil {
		log.Printf("Error creating user: %s", err)
		return
	}
	payload := JsonUser{
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
  return
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

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	var data PostUser
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
  fmt.Println(data)
	user, err := cfg.db.GetUser(r.Context(), data.Email)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	err = auth.CheckPasswordHash(data.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
	returnUser := JsonUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}
	writeResponse(w, 200, returnUser)
  return
}
