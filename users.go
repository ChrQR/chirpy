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
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
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
 
func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
  refresh, err := auth.GetBearerToken(r.Header)
  if err != nil {
    respondWithError(w, 401, "No refresh token in headers.")
    return
  }
  selectRefresh, err := cfg.db.GetRefreshToken(r.Context(), refresh)
  if err != nil {
    respondWithError(w, 401, "No refresh token found in db.")
    return
  }
  if !selectRefresh.RevokedAt.IsZero() {
    respondWithError(w, 401, "Your refresh token has expired")
    return
  }
}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type login struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var data login
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}
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
	fmt.Print(data)
	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(cfg.tokenExpiry))
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("error creating token: %v", err))
		return
	}
  refresh, err := auth.MakeRefreshToken()
  err = cfg.db.InsertRefreshToken(r.Context(), database.InsertRefreshTokenParams{
    Token: refresh,
    UserID: user.ID,
    ExpiresAt: time.Now().Add(cfg.resetExpiry),
    RevokedAt: time.Time{},
  })
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("error creating refresh_token: %v", err))
		return
	}
	returnUser := JsonUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
    RefreshToken: refresh,
	}
	writeResponse(w, 200, returnUser)
	return
}
