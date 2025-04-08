package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := "8080"

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/", fileServer)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("server running on port: %v\n", port)

	log.Fatal(server.ListenAndServe())
}
