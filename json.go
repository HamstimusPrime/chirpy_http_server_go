package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func respondWithError(w http.ResponseWriter, errMsg string, HTTPstatus int) {
	w.WriteHeader(HTTPstatus)
	errJSON, err := json.Marshal(errorMsg{Error: errMsg})
	if err != nil {
		log.Fatal("unable to parse error JSON")
	}
	w.Write([]byte(errJSON))
}
