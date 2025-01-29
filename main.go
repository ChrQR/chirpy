package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"gitea.rannes.dev/christian/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
	tokenExpiry    time.Duration
	resetExpiry    time.Duration
}

const PORT = "8080"

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("No DB_URL in .env")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("There was an error connecting to the database: %s", err)
	}

	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	dbQueries := database.New(db)
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       os.Getenv("PLATFORM"),
		secret:         os.Getenv("SECRET"),
		tokenExpiry:    1 * time.Hour,
		resetExpiry:    60 * 24 * time.Hour,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleResetUsers)
	mux.HandleFunc("GET /api/healthz", HandleHealthz)
	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	mux.HandleFunc("POST /api/refresh", apiCfg.handleRefreshToken)
	mux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirpList)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.handleGetChirp)
	log.Printf("Server listening on port %s", PORT)
	log.Fatal(srv.ListenAndServe())
}
