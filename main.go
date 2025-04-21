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

	"github.com/HamstimusPrime/chirpy_http_server_go/internal/auth"
	"github.com/HamstimusPrime/chirpy_http_server_go/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type errorMsg struct {
	Error string `json:"error"`
}

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	PLATFORM       string
}

type reqestBody struct {
	Body     string    `json:"body"`
	Email    string    `json:"email"`
	UserID   uuid.UUID `json:"user_id"`
	Password string    `json:"password"`
}

type user struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirps(w http.ResponseWriter, r *http.Request) {
	newReqBody, err := parseReqBody(r, reqestBody{})
	if err != nil {
		fmt.Printf("unable to parse request body, err: %v\n", err)
		return
	}

	chirpParams := database.CreateChirpParams{
		ID:     uuid.New(),
		Body:   newReqBody.Body,
		UserID: newReqBody.UserID,
	}

	dbChirp, err := cfg.DB.CreateChirp(context.Background(), chirpParams)
	if err != nil {
		fmt.Printf("unable to create new chirp, err: %v\n", err)
	}

	newChirp := chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		UserID:    dbChirp.UserID,
	}

	respondWithJSON(w, newChirp, http.StatusCreated)

}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	newReqBody, err := parseReqBody(r, reqestBody{})
	if err != nil {
		fmt.Printf("unable to parse request body, err: %v\n", err)
		return
	}
	//check if body contains password
	if newReqBody.Password == "" {
		responseMsg := "no password provided. Please input password"
		respondWithError(w, responseMsg, http.StatusBadRequest)
		return
	}
	//hash the password sent
	hashPassword, err := auth.HashPassword(newReqBody.Password)
	if err != nil {
		fmt.Printf("unable to hash password, err: %v\n", err)
		return
	}
	createUserParams := database.CreateUserParams{
		Email:          newReqBody.Email,
		HashedPassword: hashPassword,
	}
	dbUser, err := cfg.DB.CreateUser(context.Background(), createUserParams)
	if err != nil {
		fmt.Printf("unable to create new user, err: %v\n", err)
		return
	}

	newUser := user{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	httpResponseStatus := http.StatusCreated
	respondWithJSON(w, newUser, httpResponseStatus)
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	dbAllChirps, err := cfg.DB.GetAllChirps(context.Background())
	if err != nil {
		fmt.Printf("unable to fetcha all users, err: %v\n", err)
		return
	}
	//parse each user into a json compatible struct
	allChirps := make([]chirp, len(dbAllChirps))
	for i := range dbAllChirps {
		allChirps[i].ID = dbAllChirps[i].ID
		allChirps[i].CreatedAt = dbAllChirps[i].CreatedAt
		allChirps[i].UpdatedAt = dbAllChirps[i].UpdatedAt
		allChirps[i].Body = dbAllChirps[i].Body
	}
	respondWithJSON(w, allChirps, http.StatusOK)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDstr := r.PathValue("chirpID")

	chirpID, err := uuid.Parse(chirpIDstr)
	if err != nil {
		errorMsg := fmt.Sprintf("invalid UUID: %v, err: %v", chirpIDstr, err)
		respondWithError(w, errorMsg, http.StatusBadRequest)
		return
	}
	dbChirp, err := cfg.DB.GetChirp(context.Background(), chirpID)
	if err != nil {
		errorMsg := fmt.Sprintf("could not find user with id: %v, err: %v", chirpIDstr, err)
		respondWithError(w, errorMsg, http.StatusNotFound)
	}
	chirp := chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
	respondWithJSON(w, chirp, http.StatusOK)
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	newReqBody, err := parseReqBody(r, reqestBody{})
	if err != nil {
		errMsg := fmt.Sprintf("unable to parse request body, err: %v", err)
		fmt.Println(errMsg)
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}

	userPassword := newReqBody.Password
	userEmail := newReqBody.Email
	//check if password was provided
	if userPassword == "" {
		errMsg := "no password provided"
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}
	//check if email exists
	if userEmail == "" {
		errMsg := "no email provided"
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}

	//fetch user with email
	dbUser, err := fetchUserWithEmail(userEmail, cfg)
	if err != nil {
		errMsg := fmt.Sprintf("unable to validate user with email: %v, error: %v\n", userEmail, err)
		respondWithError(w, errMsg, http.StatusBadRequest)
	}

	//check if password of userPassword matches userPassword
	err = auth.CheckPasswordHash(dbUser.HashedPassword, userPassword)
	if err != nil {
		errorMsg := "Incorrect email or password"
		respondWithError(w, errorMsg, http.StatusUnauthorized)
		return
	}
	newUser := user{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	respondWithJSON(w, newUser, http.StatusOK)
}

func (cfg *apiConfig) handlerResetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits = atomic.Int32{}
	//check if Platform is set to dev in order to give "delete" acess to users table
	if cfg.PLATFORM != "dev" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusForbidden)
		message := "unauthorized access!"
		w.Write([]byte(message))
		return
	}
	cfg.DB.DeleteAllUsers(context.Background())
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	metricsPageHTML := fmt.Sprintf("<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>", cfg.fileserverHits.Load())
	w.Write([]byte(metricsPageHTML))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
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
	apiConfiguration := apiConfig{DB: dbQueries, fileserverHits: atomic.Int32{}, PLATFORM: platform}
	handler := http.StripPrefix("/app", fileServer)
	mux.Handle("/app/", apiConfiguration.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfiguration.handlerGetChirp)
	mux.HandleFunc("GET /api/chirps", apiConfiguration.handlerGetAllChirps)
	mux.HandleFunc("GET /admin/metrics", apiConfiguration.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiConfiguration.handlerResetMetrics)
	mux.HandleFunc("POST /api/chirps", apiConfiguration.handlerCreateChirps)
	mux.HandleFunc("POST /api/login", apiConfiguration.handlerUserLogin)
	mux.HandleFunc("POST /api/users", apiConfiguration.handlerCreateUser)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("server running on port: %v\n", port)

	log.Fatal(server.ListenAndServe())

}
