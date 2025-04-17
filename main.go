package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/HamstimusPrime/chirpy_http_server_go/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	PLATFORM       string
}

type reqestBody struct {
	Body  string `json:"body"`
	Email string `json:"email"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	metricsPageHTML := fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", cfg.fileserverHits.Load())
	w.Write([]byte(metricsPageHTML))
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits = atomic.Int32{}
	//check if Platform is set to dev in order to give delete acess to users table
	if cfg.PLATFORM != "dev" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		message := "unauthorized access!"
		w.Write([]byte(message))
		return
	}
	cfg.DB.DeleteAllUsers(context.Background())
}

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	newReqBody, err := parseReqBody(w, r, reqestBody{})
	if err != nil {
		fmt.Printf("unable to parse request body, err: %v\n", err)
	}
	dbUser, err := cfg.DB.CreateUser(context.Background(), newReqBody.Email)
	if err != nil {
		log.Fatalf("unable to create new user, err: %v\n", err)
	}

	newUser := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	httpResponseStatus := http.StatusCreated
	respondWithJSON(w, newUser, httpResponseStatus)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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
	//#comment:S+342*&
	apiConfiguration := apiConfig{DB: dbQueries, fileserverHits: atomic.Int32{}, PLATFORM: platform}
	handler := http.StripPrefix("/app", fileServer)
	mux.Handle("/app/", apiConfiguration.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiConfiguration.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiConfiguration.resetMetricsHandler)
	mux.HandleFunc("POST /api/validate_chirp", chirpValidateHandler)
	mux.HandleFunc("POST /api/users", apiConfiguration.usersHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("server running on port: %v\n", port)

	log.Fatal(server.ListenAndServe())

}
