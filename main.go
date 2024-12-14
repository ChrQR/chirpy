package main

import (
	"log"
	"net/http"

	"gitea.rannes.dev/christian/chirpy/validation"
	"gitea.rannes.dev/christian/chirpy/healthz"
	"gitea.rannes.dev/christian/chirpy/metrics"
)

const PORT = "8080"

func main() {
	mux := http.NewServeMux()
	srv := http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	apiCfg := metrics.ApiConfig{}

	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.HandlerMetricsReset)
	mux.HandleFunc("GET /api/healthz", healthz.HandleHealthz)
	mux.HandleFunc("POST /api/validate_chirp", validation.ValidateChirp)
	log.Printf("Server listening on port %s", PORT)
	log.Fatal(srv.ListenAndServe())
}
