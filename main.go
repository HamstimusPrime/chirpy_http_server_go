package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/HamstimusPrime/chirpy_http_server_go/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	jwt_secret := os.Getenv("JWT_SECRET")
	dbURL := os.Getenv("DB_URL")
	port := os.Getenv("PORT")
	platform := os.Getenv("PLATFORM")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("unable to establish connection to database: %v", err)
	}

	dbQueries := database.New(db)
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("."))

	apiConfiguration := apiConfig{DB: dbQueries,
		fileserverHits: atomic.Int32{},
		PLATFORM:       platform,
		JWT_SECRET:     jwt_secret}

	handler := http.StripPrefix("/app", fileServer)
	mux.Handle("/app/", apiConfiguration.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfiguration.handlerGetChirp)
	mux.HandleFunc("GET /api/chirps", apiConfiguration.handlerGetAllChirps)
	mux.HandleFunc("GET /admin/metrics", apiConfiguration.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiConfiguration.handlerResetMetrics)
	mux.HandleFunc("POST /api/chirps", apiConfiguration.handlerCreateChirps)
	mux.HandleFunc("POST /api/login", apiConfiguration.handlerUserLogin)
	mux.HandleFunc("POST /api/refresh", apiConfiguration.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiConfiguration.handlerRevokeToken)
	mux.HandleFunc("POST /api/users", apiConfiguration.handlerCreateUser)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("server running on port: %v\n", port)

	log.Fatal(server.ListenAndServe())

}
