package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type errorMsg struct {
	Error string `json:"error"`
}

var profaneWords = map[string]bool{
	"kerfuffle": true,
	"sharbert":  true,
	"fornax":    true,
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	message := "OK"
	w.Write([]byte(message))
}

func chirpHandler(w http.ResponseWriter, r *http.Request) {

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

	profaneWord, err := containsProfaneWords(parsedRequest.Body)
	if err != nil {
		fmt.Print(err)
		errMsg := fmt.Sprintf("request contains profane word: %v", profaneWord)
		status := http.StatusBadRequest
		respondWithError(w, errMsg, status)
		return
	}

	// filteredChirp := filterProfaneWords(parsedRequest.Body)

	// status := http.StatusOK
	// reqestBody := struct {
	// 	Cleaned_body string `json:"cleaned_body"`
	// }{
	// 	Cleaned_body: filteredChirp,
	// }

	// respondWithJSON(w, reqestBody, status)

}

func filterProfaneWords(words string) string {
	censorCharacter := "****"
	wordsList := strings.Split(words, " ")
	for i, _ := range wordsList {
		wordLower := strings.ToLower(wordsList[i])
		_, ok := profaneWords[wordLower]
		if ok {
			wordsList[i] = censorCharacter
		}
	}
	return strings.Join(wordsList, " ")
}

func containsProfaneWords(words string) (string, error) {
	wordsList := strings.Split(words, " ")
	for i, _ := range wordsList {
		wordLower := strings.ToLower(wordsList[i])
		_, ok := profaneWords[wordLower]
		if ok {
			return wordsList[i], fmt.Errorf("found profane word \n", wordsList[i])
		}
	}
	return "", nil
}
