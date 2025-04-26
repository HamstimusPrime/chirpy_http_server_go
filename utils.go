package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/HamstimusPrime/chirpy_http_server_go/internal/database"
)

func parseReqBody(req *http.Request, format reqestBody) (reqestBody, error) {
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&format)
	if err != nil {
		return reqestBody{}, err
	}
	return format, nil
}

func respondWithError(w http.ResponseWriter, errMsg string, HTTPstatus int) {
	w.WriteHeader(HTTPstatus)
	errJSON, err := json.Marshal(errorMsg{Error: errMsg})
	if err != nil {
		log.Fatal("unable to parse error JSON")
	}
	w.Write([]byte(errJSON))
}

func respondWithJSON(w http.ResponseWriter, resTemplate interface{}, HTTPstatus int) {
	resJSON, err := json.Marshal(resTemplate)
	if err != nil {
		log.Fatal("unable to parse response JSON")
	}
	w.Header().Set("Content-Type", "json/plain; charset=utf-8")
	w.WriteHeader(HTTPstatus)
	w.Write([]byte(resJSON))
}

func fetchUserWithEmail(email string, cfg *apiConfig) (database.User, error) {
	user, err := cfg.DB.GetUserByEmail(context.Background(), email)
	if err != nil {
		return database.User{}, err
	}
	return user, nil
}

func passwordInRequestBody(body reqestBody) bool {
	password := body.Password
	return password != ""
}
