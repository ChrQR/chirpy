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
  db *database.Queries
}

const PORT = "8080"

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("There was an error connecting to the database: %s", err)
	}

	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	apiCfg := apiConfig{
    fileserverHits: atomic.Int32{},
    db: database.New(db),
  }

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerMetricsReset)
	mux.HandleFunc("GET /api/healthz", HandleHealthz)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)
	log.Printf("Server listening on port %s", PORT)
	log.Fatal(srv.ListenAndServe())
}
