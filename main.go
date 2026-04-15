package main

import (
	"log"
	"net/http"
	"runtime"
	"time"

	"redis/internal/api"
	ownmap "redis/internal/map"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	om := ownmap.NewOwnMap(1_000_000)

	mux := http.NewServeMux()
	mux.HandleFunc("/set", logging(api.SetHandler(om)))
	mux.HandleFunc("/get", logging(api.GetHandler(om)))
	mux.HandleFunc("/docs", api.DocsHandler())
	mux.HandleFunc("/swagger.yaml", api.SwaggerHandler())

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("starting on :8080 with %d CPUs", runtime.NumCPU())
	log.Fatal(server.ListenAndServe())
}

func logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	}
}
