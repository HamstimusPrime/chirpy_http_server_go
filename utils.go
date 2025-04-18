package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func parseReqBody(w http.ResponseWriter, req *http.Request, format reqestBody) (reqestBody, error) {
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&format)
	if err != nil {
		errMsg := "something went wrong"
		status := http.StatusCreated
		respondWithError(w, errMsg, status)
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

// func validateUserWithEmail(email string, cfg *apiConfig) error {
// 	_, err := cfg.DB.GetUserByEmail(context.Background(), email)
// 	if err != nil {
// 		fmt.Printf("unable to validate user with email: %v, error: %v\n", email, err)
// 		return err
// 	}
// 	return nil
// }

// func validateUserWithID(cfg *apiConfig, id uuid.UUID) error {
// 	_, err := cfg.DB.GetUserByID(context.Background(), id)
// 	if err != nil {
// 		fmt.Printf("unable to validate user with email: %v, error: %v\n", id, err)
// 		return err
// 	}
// 	return nil
// }
