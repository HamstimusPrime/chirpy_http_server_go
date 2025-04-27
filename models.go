package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/HamstimusPrime/chirpy_http_server_go/internal/auth"
	"github.com/HamstimusPrime/chirpy_http_server_go/internal/database"
	"github.com/google/uuid"
)

type errorMsg struct {
	Error string `json:"error"`
}

type chirpsJSON []struct {
	Body string `json:"body"`
}

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	PLATFORM       string
	JWT_SECRET     string
	POLKA_KEY      string
}

type reqestBody struct {
	Body             string    `json:"body"`
	Email            string    `json:"email"`
	UserID           uuid.UUID `json:"user_id"`
	Password         string    `json:"password"`
	ExpiresInSeconds int       `json:"expires_in_seconds"`
}

type user struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

type chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type webhookRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (cfg *apiConfig) handlerCreateChirps(w http.ResponseWriter, r *http.Request) {
	newReqBody, err := parseReqBody(r, reqestBody{})
	if err != nil {
		fmt.Printf("unable to parse request body, err: %v\n", err)
		return
	}

	//validate user creating chirp. To create a chirp, a user needs a valid JWT
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, fmt.Sprintf("%v", err), http.StatusBadRequest)
		return
	}

	jwt_secret := cfg.JWT_SECRET
	userID, err := auth.ValidateJWT(bearerToken, jwt_secret)
	if err != nil {
		errorMsg := "error validating user"
		fmt.Printf("%s\nerror:%v\nToken : %v", errorMsg, err, bearerToken)
		respondWithError(w, errorMsg, http.StatusUnauthorized)
		return
	}

	chirpParams := database.CreateChirpParams{
		ID:     uuid.New(),
		Body:   newReqBody.Body,
		UserID: userID,
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
		Body:      dbChirp.Body,
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

	newUser := struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	}
	httpResponseStatus := http.StatusCreated
	respondWithJSON(w, newUser, httpResponseStatus)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	//get access token
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Printf("error fetching token, Token: %v\nerr: %v\n", token, err)
		errMsg := "something went wrong fetching token"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}
	//authenticate user using JWT and get userID linked to token
	userID, err := auth.ValidateJWT(token, cfg.JWT_SECRET)
	if err != nil {
		fmt.Printf("error validating token, Token: %v\nerr: %v\n", token, err)
		errMsg := "unauthorized access"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}

	chirpIDstr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDstr)
	if err != nil {
		errorMsg := fmt.Sprintf("invalid UUID: %v, err: %v", chirpIDstr, err)
		respondWithError(w, errorMsg, http.StatusBadRequest)
		return
	}

	//check if chirp exists
	_, err = cfg.DB.GetChirp(context.Background(), chirpID)
	if err != nil {
		errorMsg := fmt.Sprintf("error fetching chirp with chirpID: %v\n, err: %v\n", chirpID, err)
		respondWithError(w, errorMsg, http.StatusNotFound)
		return
	}

	getChirpByUserIDparams := database.GetChirpByUserIDParams{
		UserID: userID,
		ID:     chirpID,
	}
	_, err = cfg.DB.GetChirpByUserID(context.Background(), getChirpByUserIDparams)
	if err != nil {
		errorMsg := fmt.Sprintf("error matching userID: %v, with chirpID: %v, err: %v\n", userID, chirpID, err)
		respondWithError(w, errorMsg, http.StatusForbidden)
		return
	}

	//delete chirp
	err = cfg.DB.DeleteChirpByID(context.Background(), chirpID)
	if err != nil {
		errorMsg := "something went wrong deleting chirp"
		respondWithError(w, errorMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	//check for author ID in query string from client and parse to uuid
	authorIDstr := r.URL.Query().Get("author_id")
	if authorIDstr != "" {
		authorID, err := uuid.Parse(authorIDstr)
		if err != nil {
			fmt.Printf("failed to parse authorID from client\nauthorID: %v\nerr: %v\n", authorID, err)
			errorMsg := "something went wrong deleting chirp"
			respondWithError(w, errorMsg, http.StatusInternalServerError)
			return
		}

		//fetch chirps using parsed author ID
		dbChirps, err := cfg.DB.GetChirpsByUserID(context.Background(), authorID)
		if err != nil {
			fmt.Printf("failed to fetch chirps with authorID\nauthorID: %v\nerr: %v\n", authorID, err)
			errorMsg := "something went wrong deleting chirp"
			respondWithError(w, errorMsg, http.StatusBadRequest)
			return
		}

		//populate chirpJSON with chirps
		chirps := make(chirpsJSON, len(dbChirps))
		for i := range dbChirps {
			chirps[i].Body = dbChirps[i]
		}
		respondWithJSON(w, chirps, http.StatusOK)
		return
	}

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

func (cfg *apiConfig) handlerGetChirpWithID(w http.ResponseWriter, r *http.Request) {
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
		return
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

func (cfg *apiConfig) handlerUpdatePassword(w http.ResponseWriter, r *http.Request) {
	newReqBody, err := parseReqBody(r, reqestBody{})
	if err != nil {
		errMsg := fmt.Sprintf("unable to parse request body, err: %v\n", err)
		fmt.Println(errMsg)
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}
	//get token from client
	token, err := auth.GetBearerToken(r.Header)
	errMsg := "something went wrong"
	if err != nil {
		fmt.Printf("unable to get bearer token, err: %v\n", err)
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}
	//check for password from request.
	if !passwordInRequestBody(newReqBody) {
		errMsg := "no password provided"
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}
	//fetch user ID using jwt token
	userID, err := auth.ValidateJWT(token, cfg.JWT_SECRET)
	if err != nil {
		fmt.Printf("error validating user with token:%v \nerr: %v\n", token, err)
		errMsg := "something went wrong"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}

	dbUser, err := cfg.DB.GetUserByID(context.Background(), userID)
	if err != nil {
		fmt.Printf("error fetching user with token:%v \nerr: %v\n", token, err)
		errMsg := "unauthorized access"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}
	//hash password sent from client
	hashedPassword, err := auth.HashPassword(newReqBody.Password)
	if err != nil {
		fmt.Printf("error hashing password:%v \nerr: %v\n", token, err)
		errMsg := "something went wrong"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	//update email of the user
	updateEmailParams := database.UpdateUserEmailParams{
		Email:     newReqBody.Email,
		ID:        dbUser.ID,
		UpdatedAt: time.Now(),
	}

	//update password of user
	updatePasswordParams := database.UpdateUserPasswordParams{
		HashedPassword: hashedPassword,
		ID:             dbUser.ID,
	}
	err = cfg.DB.UpdateUserPassword(context.Background(), updatePasswordParams)
	if err != nil {
		fmt.Printf("error updating password with user:%v \nerr: %v\n", dbUser.ID, err)
		errMsg := "unauthorized access"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}

	//updated user here
	userData, err := cfg.DB.UpdateUserEmail(context.Background(), updateEmailParams)
	if err != nil {
		fmt.Printf("error updating emial with user:%v \nerr: %v\n", dbUser.ID, err)
		errMsg := "unauthorized access"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}

	updatedUser := struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}{
		ID:          dbUser.ID,
		CreatedAt:   userData.UpdatedAt,
		UpdatedAt:   userData.UpdatedAt,
		Email:       userData.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	}

	respondWithJSON(w, updatedUser, http.StatusOK)

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
	expiry_time := newReqBody.ExpiresInSeconds

	//check if password was provided
	if !passwordInRequestBody(newReqBody) {
		errMsg := "no password provided"
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}

	//check if email was provided
	if userEmail == "" {
		errMsg := "no email provided"
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}

	//check if expiry time was provided. Set default expiry if none provided
	default_expiry_time := 3600
	if expiry_time == 0 {
		expiry_time = default_expiry_time
	}
	if expiry_time > default_expiry_time {
		expiry_time = default_expiry_time
	}

	//fetch user with email provided
	dbUser, err := fetchUserWithEmail(userEmail, cfg)
	if err != nil {
		errMsg := fmt.Sprintf("unable to validate user with email: %v, error: %v\n", userEmail, err)
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}

	//check if password provided matches Hased passsword record of user in DB
	err = auth.CheckPasswordHash(dbUser.HashedPassword, userPassword)
	if err != nil {
		errorMsg := "Incorrect email or password"
		respondWithError(w, errorMsg, http.StatusUnauthorized)
		return
	}

	//generate JWT for successfully logged in user
	jwtToken, err := auth.MakeJWT(dbUser.ID, cfg.JWT_SECRET, time.Duration(expiry_time)*time.Second)
	if err != nil {
		errMsg := "unable to generate token"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	//generate refresh token for user and make entry into refresh_tokens table
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		fmt.Printf("unable to generate refresh token")
		errMsg := "something went wrong"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}
	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour), //60 days
		RevokedAt: sql.NullTime{Valid: false},
	}
	err = cfg.DB.CreateRefreshToken(context.Background(), refreshTokenParams)
	if err != nil {
		fmt.Printf("unable to create refresh token")
		errMsg := "something went wrong"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	newUser := user{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        jwtToken,
		RefreshToken: refreshToken,
		IsChirpyRed:  dbUser.IsChirpyRed,
	}
	respondWithJSON(w, newUser, http.StatusOK)
}

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	//return 404 if key from client does not match server API key
	clientKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		fmt.Printf("client key mismatch\nkey: %v\n", clientKey)
		errMsg := "unauthorized access"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}

	hookRequestData, err := parseHookRequestBody(r, webhookRequest{})
	if err != nil {
		errMsg := fmt.Sprintf("unable to parse hook request body, err: %v", err)
		fmt.Println(errMsg)
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}

	if hookRequestData.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	//update is_chirpy_red in DB and return 204 if ID sent by client isn't found
	userID, err := uuid.Parse(hookRequestData.Data.UserID)
	if err != nil {
		fmt.Printf("unable to parse uuid: %v\n", hookRequestData.Data.UserID)
		errMsg := "something went wrong"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	isChirpyRedParams := database.UpdateIsChirpyRedParams{
		IsChirpyRed: true,
		ID:          userID,
	}

	err = cfg.DB.UpdateIsChirpyRed(context.Background(), isChirpyRedParams)

	if err != nil {
		fmt.Printf("unable to update is_chirpy_red, err:%v\n", err)
		errMsg := "something went wrong"
		respondWithError(w, errMsg, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)

}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	errMsg := "something went wrong"
	if err != nil {
		fmt.Printf("unable to get bearer token, err: %v", err)
		respondWithError(w, errMsg, http.StatusBadRequest)
		return
	}

	//check if token from client exists in refresh token table
	dbUser, err := cfg.DB.GetUserFromRefreshToken(context.Background(), token)
	if err != nil {
		fmt.Printf("error getting user with token: %v\n, err: %v\n", token, err)
		errMsg := "unable to refresh token"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}

	//check if token is expired
	if time.Now().After(dbUser.ExpiresAt) {
		fmt.Printf("epired token, token: %v\n", dbUser.ExpiresAt)
		errMsg := "expired token"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}

	//check if token has been revoked
	if dbUser.RevokedAt.Valid {
		fmt.Printf("revoked token, revoked at: %v\n", dbUser.RevokedAt)
		errMsg := "revoked token"
		respondWithError(w, errMsg, http.StatusUnauthorized)
		return
	}

	//respond with new JWT that expires in an hour if token exists
	jwt_token, err := auth.MakeJWT(dbUser.ID, cfg.JWT_SECRET, time.Hour)
	if err != nil {
		fmt.Printf("unable to generate JWT, err: %v", err)
		errMsg := "something went wrong"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	resBody := struct {
		Token string `json:"token"`
	}{
		Token: jwt_token,
	}
	respondWithJSON(w, resBody, http.StatusOK)

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
	err := cfg.DB.DeleteAllUsers(context.Background())
	if err != nil {
		errMsg := "unable to reset all user entries"
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	errMsg := "something went wrong"

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		fmt.Printf("unable to get bearer token, err: %v", err)
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	err = cfg.DB.RevokeToken(context.Background(), token)
	if err != nil {
		fmt.Printf("unable to revoke token, err: %v", err)
		respondWithError(w, errMsg, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
