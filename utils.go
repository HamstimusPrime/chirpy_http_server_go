package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func parseReqBody(w http.ResponseWriter, req *http.Request, format reqestBody) (reqestBody, error) {
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&format)
	if err != nil {
		errMsg := fmt.Sprintf("something went wrong, err: %v\n", err)
		status := http.StatusBadRequest
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
