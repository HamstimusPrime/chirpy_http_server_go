package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middleWareWriteMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	hitRequest := cfg.fileserverHits
	numRequest := fmt.Sprintf("Hits: %d", hitRequest.Load())
	w.Write([]byte(numRequest))
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	cfg.fileserverHits = atomic.Int32{}
}

func main() {
	port := "8080"

	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("."))

	apiConfiguration := apiConfig{}
	handler := http.StripPrefix("/app", fileServer)
	mux.Handle("/app/", apiConfiguration.middlewareMetricsInc(handler))
	mux.HandleFunc("/healthz", readinessHandler)
	mux.HandleFunc("/metrics", apiConfiguration.metricsHandler)
	mux.HandleFunc("/reset", apiConfiguration.resetMetricsHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("server running on port: %v\n", port)

	log.Fatal(server.ListenAndServe())

}
