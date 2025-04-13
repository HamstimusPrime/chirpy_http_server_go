package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type reqestBody struct {
	Body string `json:"body"`
}

type errorMsg struct {
	Error string `json:"error"`
}

type resMsg struct {
	Valid bool `json:"valid"`
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	message := "OK"
	w.Write([]byte(message))
}

func chirpValidateHandler(w http.ResponseWriter, r *http.Request) {

	parsedRequest := reqestBody{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&parsedRequest)

	if err != nil {
		errMsg := "something went wrong"
		status := http.StatusCreated
		respondWithError(w, errMsg, status)
		return
	}

	if len(parsedRequest.Body) > 140 {
		errMsg := "Chirp is too long"
		status := http.StatusBadRequest
		respondWithError(w, errMsg, status)
		return
	}

	resJSON, err := json.Marshal(resMsg{Valid: true})
	if err != nil {
		log.Fatal("unable to parse response JSON")
	}
	w.Header().Set("Content-Type", "json/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(resJSON))
}
