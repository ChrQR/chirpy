package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"gitea.rannes.dev/christian/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
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
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handleResetUsers)
	mux.HandleFunc("GET /api/healthz", HandleHealthz)
	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	mux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirpList)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.handleGetChirp)
	log.Printf("Server listening on port %s", PORT)
	log.Fatal(srv.ListenAndServe())
}

